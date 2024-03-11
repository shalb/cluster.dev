package project

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/apex/log"
	"github.com/paulrademacher/climenu"
	"github.com/shalb/cluster.dev/pkg/colors"
	"github.com/shalb/cluster.dev/pkg/config"
	"github.com/shalb/cluster.dev/pkg/utils"
)

// Build generate all terraform code for project.
func (p *Project) Build() error {
	err := p.ClearCacheDir()
	if err != nil {
		return fmt.Errorf("build project: %v", err.Error())
	}
	for _, unit := range p.Units {
		if err := unit.Build(); err != nil {
			return err
		}
	}

	return nil
}

// Destroy all units.
func (p *Project) Destroy() error {
	planStatus := ProjectPlanningStatus{}
	p.planDestroyAll(&planStatus)
	destroyGraph, err := planStatus.BuildGraph()
	if err != nil {
		return fmt.Errorf("build destroy graph: %w", err)
	}
	if destroyGraph.Len() < 1 {
		log.Info("Nothing to destroy, exiting")
		return nil
	}
	if !config.Global.Force {
		stopChan := make(chan struct{})
		p.StartSigTrap(stopChan)
		showPlanResults(destroyGraph)
		if p.NewVersionMessage != "" {
			log.Info(p.NewVersionMessage)
		}
		respond := climenu.GetText("Continue?(yes/no)", "no")
		stopChan <- struct{}{}
		if respond != "yes" {
			log.Info("Destroying cancelled")
			return nil
		}
	}
	err = p.ClearCacheDir()
	if err != nil {
		return fmt.Errorf("project destroy: clear cache dir: %w", err)
	}
	log.Info("Destroying...")
	for {
		// log.Warnf("FOR Project apply. Unit links: %+v", p.UnitLinks)
		if destroyGraph.Len() == 0 {
			return p.OwnState.SaveState()
		}
		gUnit, fn, err := destroyGraph.GetNextAsync()
		if err != nil {
			unitName := ""
			if gUnit != nil {
				unitName = gUnit.UnitPtr.Key()
			}
			log.Errorf("error in unit %v, waiting for all running units done.", unitName)
			destroyGraph.Wait()
			for _, e := range destroyGraph.Errors() {
				log.Errorf("unit: '%v':\n%v", e.Key(), e.ExecError())
			}
			err := p.OwnState.SaveState()
			if err != nil {
				return fmt.Errorf("save state after error: %w", err)
			}
			return fmt.Errorf("applying error")
		}
		// Check if graph return nil unit - applying finished, return
		if gUnit == nil {
			return p.OwnState.SaveState()
		}
		switch gUnit.Operation {
		case Apply, Update:
			// TODO remove this check
			return fmt.Errorf("destroy: internal error, found unit for apply in destroy command")
		case Destroy:
			// log.Warnf("DESTROY circle: run DESTROY for unit: %v", gUnit.UnitPtr.Key())
			go destroyRoutine(gUnit, fn, p)
		}
	}
}

func (p *Project) StartSigTrap(stop chan struct{}) {
	p.HupUnlockChan = make(chan os.Signal, 1)
	signals := []os.Signal{syscall.SIGTERM, syscall.SIGINT}
	signal.Notify(p.HupUnlockChan, signals...)
	go func() {
		for {
			select {
			case <-p.HupUnlockChan:
				fmt.Println()
				log.Warnf("Execution halted. Unlocking state and exiting...")
				p.UnLockState()
				close(p.HupUnlockChan)
				os.Exit(0)
			case <-stop:
				close(p.HupUnlockChan)
				return
			}
		}
	}()
}

func (p *Project) StopSigTrap() {
	close(p.HupUnlockChan)
}

// Apply all units.
func (p *Project) Apply() error {
	// var applyGraph *ProjectPlanningStatus
	applyGraph, err := p.Plan()
	if err != nil {
		return err
	}
	if !config.Global.Force {
		if !applyGraph.planningUnits.HasChanges() {
			return nil
		}
		if p.NewVersionMessage != "" {
			log.Info(p.NewVersionMessage)
		}
		stopChan := make(chan struct{})
		p.StartSigTrap(stopChan)
		respond := climenu.GetText("Continue?(yes/no)", "no")
		stopChan <- struct{}{}
		if respond != "yes" {
			log.Info("Cancelled")
			return nil
		}
	}
	err = p.ClearCacheDir()
	if err != nil {
		return fmt.Errorf("project apply: clear cache dir: %v", err.Error())
	}
	log.Info("Applying...")

	for {
		// log.Warnf("FOR Project apply. Unit links: %+v", p.UnitLinks)
		if applyGraph.Len() == 0 {
			return p.OwnState.SaveState()
		}
		gUnit, fn, err := applyGraph.GetNextAsync()
		if err != nil {
			unitName := ""
			if gUnit != nil {
				unitName = gUnit.UnitPtr.Key()
			}
			log.Errorf("error in unit %v, waiting for all running units done.", unitName)
			applyGraph.Wait()
			for _, e := range applyGraph.Errors() {
				log.Errorf("unit: '%v':\n%v", e.Key(), e.ExecError())
			}
			err := p.OwnState.SaveState()
			if err != nil {
				return fmt.Errorf("save state after error: %w", err)
			}
			return fmt.Errorf("applying error")
		}
		// Check if graph return nil unit - applying finished, return
		if gUnit == nil {
			p.OwnState.SaveState()
			return nil
		}
		switch gUnit.Operation {
		case Apply, Update:
			// log.Warnf("APPLY circle: run APPLY for unit: %v", gUnit.UnitPtr.Key())
			go applyRoutine(gUnit, fn, p)
		case Destroy:
			// log.Warnf("APPLY circle: run DESTROY for unit: %v", gUnit.UnitPtr.Key())
			go destroyRoutine(gUnit, fn, p)
		}
	}
}

// applyRoutine function to run unit apply in parallel
func applyRoutine(graphUnit *UnitPlanningStatus, finFunc func(error), p *Project) {
	log.Infof(colors.Fmt(colors.LightWhiteBold).Sprintf("Applying unit '%v':", graphUnit.UnitPtr.Key()))
	err := graphUnit.UnitPtr.Build()
	if err != nil {
		log.Errorf("unit build error: %v", err.Error())
		finFunc(err)
		return
	}
	p.ProcessedUnitsCount++
	err = graphUnit.UnitPtr.Apply()
	if err != nil {
		state, _ := utils.JSONEncode(graphUnit.UnitPtr)
		log.Warnf("applyRoutine: %v", string(state))
		if graphUnit.UnitPtr.IsTainted() {
			// log.Warnf("applyRoutine: tainted %v", graphUnit.UnitPtr.Key())
			p.OwnState.UpdateUnit(graphUnit.UnitPtr)
		}
		finFunc(fmt.Errorf("apply unit: %v", err.Error()))
		return
	}
	err = graphUnit.UnitPtr.UpdateProjectRuntimeData(p)
	if err != nil {
		finFunc(err)
		return
	}
	p.OwnState.UpdateUnit(graphUnit.UnitPtr)
	graphUnit.UnitPtr.SetExecStatus(Finished)
	finFunc(nil)
}

// destroyRoutine function to run unit destroy in parallel
func destroyRoutine(graphUnit *UnitPlanningStatus, finFunc func(error), p *Project) {
	log.Infof(colors.Fmt(colors.LightWhiteBold).Sprintf("Destroying unit '%v':", graphUnit.UnitPtr.Key()))
	err := graphUnit.UnitPtr.Build()
	if err != nil {
		log.Errorf("project apply: unit build error: %v", err.Error())
		finFunc(err)
		return
	}
	p.ProcessedUnitsCount++
	err = graphUnit.UnitPtr.Destroy()
	if err != nil {
		finFunc(fmt.Errorf("destroy unit: %v", err.Error()))
		return
	}
	p.OwnState.DeleteUnit(graphUnit.UnitPtr)
	graphUnit.UnitPtr.SetExecStatus(Finished)
	finFunc(nil)
}

// Plan and output result.
func (p *Project) Plan() (*graph, error) {
	log.Infof(colors.Fmt(colors.LightWhiteBold).Sprintf("Checking units in state"))
	planningSt, err := p.buildPlan()
	if err != nil {
		return nil, err
	}
	showPlanResults(planningSt)
	return planningSt, nil
}

// planDestroy collect and show units for destroying.
func (p *Project) planDestroy(opStatus *ProjectPlanningStatus) {
	for _, md := range p.OwnState.UnitsSlice() {
		if i := findUnit(p.UnitsSlice(), md); i >= 0 {
			continue
		}
		diff := utils.Diff(md.GetDiffData(), nil, true)
		opStatus.Add(md, Destroy, diff, md.IsTainted())
	}
}

// planDestroyAll add all units from state for destroying.
func (p *Project) planDestroyAll(opStatus *ProjectPlanningStatus) {
	var units []Unit
	if config.Global.IgnoreState {
		units = p.UnitsSlice()
	} else {
		units = p.OwnState.UnitsSlice()
	}
	for _, md := range units {
		diff := utils.Diff(md.GetDiffData(), nil, true)
		opStatus.Add(md, Destroy, diff, md.IsTainted())
	}
}

// findUnit returns the index of unitForSearch in the list. Returns -1 if not found.
func findUnit(list []Unit, unitForSearch Unit) int {
	if len(list) < 1 {
		return -1
	}
	for index, m := range list {
		if unitForSearch.Key() == m.Key() {
			return index
		}
	}
	return -1
}

// Plan and output result.
func (p *Project) buildPlan() (resGraph *graph, err error) {
	err = checkUnitDependencies(p)
	if err != nil {
		return
	}
	planningStatus := &ProjectPlanningStatus{}
	if config.Global.IgnoreState {
		for _, u := range p.UnitsSlice() {
			planningStatus.Add(u, Apply, utils.Diff(nil, u.GetDiffData(), true), false)
		}
		return planningStatus.BuildGraph()
	}
	p.planDestroy(planningStatus)
	for _, unit := range p.UnitsSlice() {
		_, exists := p.OwnState.Units[unit.Key()]
		diff, stateUnit := p.OwnState.CheckUnitChanges(unit)
		if len(diff) > 0 || config.Global.IgnoreState {
			if len(diff) > 0 {
				if exists {
					planningStatus.Add(unit, Update, diff, stateUnit.IsTainted())
				} else {
					planningStatus.Add(unit, Apply, diff, false)
				}
			}
		} else {
			if stateUnit != nil {
				stateUnit.UpdateProjectRuntimeData(p)
			}
			planningStatus.Add(unit, NotChanged, "", false)

			// Unit was not changed. Copy unit outputs from state.
			p.UnitLinks.JoinWithDataReplace(p.OwnState.UnitLinks.ByTargetUnit(unit))
		}
	}
	// // planningStatus.Print()
	// changedUnits := planningStatus.OperationFilter(Apply, Update, Destroy)
	// for {
	// 	addedAsForceDepCount := 0
	// 	for _, st := range changedUnits.Slice() {
	// 		err = DependenciesRecursiveIterate(st.UnitPtr, func(unitForCheck Unit) error {
	// 			if unitForCheck.ForceApply() {
	// 				fu := changedUnits.FindUnit(unitForCheck)
	// 				if fu == nil {
	// 					log.Debugf("Unit '%v' added for update as a force_apply dependency", unitForCheck.Key())
	// 					if u := planningStatus.FindUnitByKey(unitForCheck); u == nil {
	// 						planningStatus.AddIfNotExists(unitForCheck, Update, colors.Fmt(colors.Yellow).Sprint("+/- Will be applied as a 'force_apply' dependency"), false)
	// 						addedAsForceDepCount++
	// 					}
	// 				}
	// 			}
	// 			return nil
	// 		})
	// 		if err != nil {
	// 			return nil, err
	// 		}
	// 	}
	// 	if addedAsForceDepCount == 0 {
	// 		break
	// 	}
	// }
	// Check graph and set sequence indexes
	resGraph, err = planningStatus.BuildGraph()
	if err != nil {
		return nil, fmt.Errorf("check apply graph: %w", err)
	}
	return
}
