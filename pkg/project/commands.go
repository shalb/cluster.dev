package project

import (
	"fmt"

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
	log.Errorf("Destroy planStatus %++v", planStatus.units)
	graph, err := planStatus.GetDestroyGraph()
	if err != nil {
		return fmt.Errorf("build destroy graph: %w", err)
	}
	if graph.Len() < 1 {
		log.Info("Nothing to destroy, exiting")
		return nil
	}
	destSeq := graph.Slice()
	log.Warnf("Destroy: destSeq: %++v", destSeq)
	if !config.Global.Force {
		showPlanResults(&planStatus)
		respond := climenu.GetText("Continue?(yes/no)", "no")
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
	for _, unit := range destSeq {
		log.Infof(colors.Fmt(colors.LightWhiteBold).Sprintf("Destroying unit '%v'", unit.UnitPtr.Key()))
		err = unit.UnitPtr.Build()
		defer unit.UnitPtr.SetExecStatus(Finished) // Set unit status done on any error
		if err != nil {
			return fmt.Errorf("project destroy: destroying deleted unit: %w", err)
		}
		p.ProcessedUnitsCount++
		err = unit.UnitPtr.Destroy()
		if err != nil {
			if unit.UnitPtr.IsTainted() {
				err = p.OwnState.SaveState()
				if err != nil {
					return fmt.Errorf("project destroy: saving state: %w", err)
				}
			}
			return fmt.Errorf("project destroy: %w", err)
		}
		p.OwnState.DeleteUnit(unit.UnitPtr)
		err = p.OwnState.SaveState()
		if err != nil {
			return fmt.Errorf("project destroy: saving state: %w", err)
		}
		unit.UnitPtr.SetExecStatus(Finished) // Set unit status done if unit destroyed without errors
	}
	return nil
}

// Apply all units.
func (p *Project) Apply() error {
	var planningStatus *ProjectPlanningStatus
	planningStatus, err := p.Plan()
	if err != nil {
		return err
	}
	if !config.Global.Force {
		if !planningStatus.HasChanges() {
			return nil
		}
		respond := climenu.GetText("Continue?(yes/no)", "no")
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
	StateDestroyGraph, err := planningStatus.GetDestroyGraph()
	if err != nil {
		return fmt.Errorf("build destroy greph: %w", err)
	}
	for _, graphUnit := range StateDestroyGraph.Slice() {
		if config.Global.IgnoreState {
			break
		}
		defer graphUnit.UnitPtr.SetExecStatus(Finished) // Set unit status done on any error
		err = graphUnit.UnitPtr.Build()
		if err != nil {
			log.Errorf("project apply: destroying deleted unit: %v", err.Error())
		}
		err = graphUnit.UnitPtr.Destroy()
		if err != nil {
			if graphUnit.UnitPtr.IsTainted() {
				err = p.OwnState.SaveState()
				if err != nil {
					return fmt.Errorf("project apply: saving state: %w", err)
				}
			}
			return fmt.Errorf("project apply: destroying deleted unit: %v", err.Error())
		}
		p.OwnState.DeleteUnit(graphUnit.UnitPtr)
		err = p.OwnState.SaveState()
		if err != nil {
			return fmt.Errorf("project apply: %v", err.Error())
		}
		graphUnit.UnitPtr.SetExecStatus(Finished)
	}
	gr, err := planningStatus.GetApplyGraph()
	if err != nil {
		return fmt.Errorf("build apply graph: %w", err)
	}
	for {
		// log.Warnf("FOR Project apply. Unit links: %+v", p.UnitLinks)
		if gr.Len() == 0 {
			p.SaveState()
			return nil
		}
		md, fn, err := gr.GetNextAsync()
		if err != nil {
			unitName := ""
			if md != nil {
				unitName = md.Key()
			}
			log.Errorf("error in unit %v, waiting for all running units done.", unitName)
			gr.Wait()
			for modKey, e := range gr.Errors() {
				log.Errorf("unit: '%v':\n%v", modKey, e.ExecError())
			}
			return fmt.Errorf("applying error")
		}
		if md == nil {
			p.SaveState()
			return nil
		}

		go func(unit Unit, finFunc func(error)) {
			log.Infof(colors.Fmt(colors.LightWhiteBold).Sprintf("Applying unit '%v':", md.Key()))
			err = unit.Build()
			if err != nil {
				log.Errorf("project apply: unit build error: %v", err.Error())
				finFunc(err)
				return
			}
			p.ProcessedUnitsCount++
			applyError := unit.Apply()
			log.Warnf("apply routine: unit done")
			if applyError != nil {
				if unit.IsTainted() {
					unit.Project().OwnState.UpdateUnit(unit)
					err := unit.Project().OwnState.SaveState()
					if err != nil {
						log.Warnf("apply routine: send sig 1")
						finFunc(err)
						return
					}
					log.Warnf("apply routine: send sig 2")
					finFunc(applyError)
					return
				}
			}
			unit.Project().OwnState.UpdateUnit(unit)
			err := unit.Project().OwnState.SaveState()
			if err != nil {
				log.Warnf("apply routine: send sig 3")
				finFunc(err)
				return
			}
			err = unit.UpdateProjectRuntimeData(p)
			if err != nil {
				log.Warnf("apply routine: send sig 4")
				finFunc(err)
				return
			}
			log.Warnf("apply routine: send sig 5")
			finFunc(nil)
		}(md, fn)
	}
}

// Plan and output result.
func (p *Project) Plan() (*ProjectPlanningStatus, error) {
	log.Infof(colors.Fmt(colors.LightWhiteBold).Sprintf("Checking units in state"))
	planningSt, err := p.buildPlan()
	if err != nil {
		return nil, err
	}
	if config.Global.ShowTerraformPlan {
		err = p.ClearCacheDir()
		if err != nil {
			return nil, fmt.Errorf("build dir for exec terraform plan: %w", err)
		}
		planSeq, err := planningSt.GetApplyGraph()
		if err != nil {
			return nil, fmt.Errorf("build graph: %w", err)
		}
		for _, md := range planSeq.Slice() {
			if err != nil {
				return nil, fmt.Errorf("project plan: clear cache dir: %v", err.Error())
			}
			allDepsDeployed := true
			for _, planModDep := range md.UnitPtr.Dependencies().Slice() {
				dS := planningSt.FindUnit(planModDep.Unit)
				if dS != nil && dS.Operation != NotChanged {
					allDepsDeployed = false
				}
			}
			if allDepsDeployed {
				err = md.UnitPtr.Build()
				if err != nil {
					log.Errorf("terraform plan: unit build error: %v", err.Error())
					return nil, err
				}
				err = md.UnitPtr.Plan()
				if err != nil {
					log.Errorf("unit '%v' terraform plan return an error: %v", md.UnitPtr.Key(), err.Error())
					return nil, err
				}
			} else {
				log.Warnf("The unit '%v' has dependencies that have not yet been deployed. Can't show terraform plan.", md.UnitPtr.Key())
			}
		}
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
	log.Errorf("planDestroyAll %++v", units)
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
func (p *Project) buildPlan() (planningStatus *ProjectPlanningStatus, err error) {
	err = checkUnitDependencies(p)
	if err != nil {
		return
	}
	planningStatus = &ProjectPlanningStatus{}
	if config.Global.IgnoreState {
		for _, u := range p.UnitsSlice() {
			planningStatus.Add(u, Apply, utils.Diff(nil, u.GetDiffData(), true), false)
		}
		return planningStatus, nil
	}
	p.planDestroy(planningStatus)
	for _, unit := range p.UnitsSlice() {
		_, exists := p.OwnState.Units[unit.Key()]
		diff, stateUnit, tainted := p.OwnState.CheckUnitChanges(unit)
		if len(diff) > 0 || config.Global.IgnoreState {
			if len(diff) > 0 {
				if exists {
					planningStatus.Add(unit, Update, diff, tainted)
				} else {
					planningStatus.Add(unit, Apply, diff, tainted)
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
	// planningStatus.Print()
	changedUnits := planningStatus.OperationFilter(Apply, Update, Destroy)
	for _, st := range changedUnits.Slice() {
		err = DependenciesRecursiveIterate(st.UnitPtr, func(unitForCheck Unit) error {
			if unitForCheck.ForceApply() {
				fu := changedUnits.FindUnit(unitForCheck)
				if fu == nil {
					log.Debugf("Unit '%v' added for update as a force_apply dependency", unitForCheck.Key())
					planningStatus.AddOrUpdate(unitForCheck, Update, colors.Fmt(colors.Yellow).Sprint("+/- Will be applied as a 'force_apply' dependency"))
				}
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	// Check graph and set sequence indexes
	_, err = planningStatus.OperationFilter(Apply, Update).GetApplyGraph()
	if err != nil {
		return nil, fmt.Errorf("check apply graph: %w", err)
	}
	return
}
