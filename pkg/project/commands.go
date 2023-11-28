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
	graph := planStatus.GetDestroyGraph()
	if graph.Len() < 1 {
		log.Info("Nothing to destroy, exiting")
		return nil
	}
	destSeq := graph.GetSequenceSet()
	if !config.Global.Force {
		showPlanResults(&planStatus)
		respond := climenu.GetText("Continue?(yes/no)", "no")
		if respond != "yes" {
			log.Info("Destroying cancelled")
			return nil
		}
	}
	err := p.ClearCacheDir()
	if err != nil {
		return fmt.Errorf("project destroy: clear cache dir: %w", err)
	}
	log.Info("Destroying...")
	for _, unit := range destSeq {
		log.Infof(colors.Fmt(colors.LightWhiteBold).Sprintf("Destroying unit '%v'", unit.Key()))
		err = unit.Build()
		if err != nil {
			return fmt.Errorf("project destroy: destroying deleted unit: %w", err)
		}
		p.ProcessedUnitsCount++
		err = unit.Destroy()
		if err != nil {
			if unit.IsTainted() {
				err = p.OwnState.SaveState()
				if err != nil {
					return fmt.Errorf("project destroy: saving state: %w", err)
				}
			}
			return fmt.Errorf("project destroy: %w", err)
		}
		p.OwnState.DeleteUnit(unit)
		err = p.OwnState.SaveState()
		if err != nil {
			return fmt.Errorf("project destroy: saving state: %w", err)
		}
	}
	return nil
}

// Apply all units.
func (p *Project) Apply() error {
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
	gr := planningStatus.GetApplyGraph()
	defer gr.Close()
	StateDestroyGraph := planningStatus.GetDestroyGraph()
	defer StateDestroyGraph.Close()

	for _, unit := range StateDestroyGraph.GetSequenceSet() {
		err = unit.Build()
		if err != nil {
			log.Errorf("project apply: destroying deleted unit: %v", err.Error())
		}
		err = unit.Destroy()
		if err != nil {
			if unit.IsTainted() {
				err = p.OwnState.SaveState()
				if err != nil {
					return fmt.Errorf("project apply: saving state: %w", err)
				}
			}
			return fmt.Errorf("project apply: destroying deleted unit: %v", err.Error())
		}
		p.OwnState.DeleteUnit(unit)
		err = p.OwnState.SaveState()
		if err != nil {
			return fmt.Errorf("project apply: %v", err.Error())
		}
	}
	for {
		// log.Warnf("FOR Project apply. Unit links: %+v", p.UnitLinks)
		if gr.Len() == 0 {
			p.SaveState()
			return nil
		}
		md, fn, err := gr.GetNextAsync()
		if err != nil {
			log.Errorf("error in unit %v, waiting for all running units done.", md.Key())
			gr.Wait()
			for modKey, e := range gr.Errors() {
				log.Errorf("unit: '%v':\n%v", modKey, e.Error())
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
			if applyError != nil {
				if unit.IsTainted() {
					unit.Project().OwnState.UpdateUnit(unit)
					err := unit.Project().OwnState.SaveState()
					if err != nil {
						finFunc(err)
						return
					}
					finFunc(applyError)
					return
				}
			}
			unit.Project().OwnState.UpdateUnit(unit)
			err := unit.Project().OwnState.SaveState()
			if err != nil {
				finFunc(err)
				return
			}
			err = unit.UpdateProjectRuntimeData(p)
			if err != nil {
				finFunc(err)
				return
			}

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
		for _, md := range planningSt.GetApplyGraph().GetSequenceSet() {
			if err != nil {
				return nil, fmt.Errorf("project plan: clear cache dir: %v", err.Error())
			}
			allDepsDeployed := true
			for _, planModDep := range md.Dependencies().Slice() {
				dS := planningSt.FindUnit(planModDep.Unit)
				if dS != nil && dS.Status != NotChanged {
					allDepsDeployed = false
				}
			}
			if allDepsDeployed {
				err = md.Build()
				if err != nil {
					log.Errorf("terraform plan: unit build error: %v", err.Error())
					return nil, err
				}
				err = md.Plan()
				if err != nil {
					log.Errorf("unit '%v' terraform plan return an error: %v", md.Key(), err.Error())
					return nil, err
				}
			} else {
				log.Warnf("The unit '%v' has dependencies that have not yet been deployed. Can't show terraform plan.", md.Key())
			}
		}
	}
	showPlanResults(planningSt)
	return planningSt, nil
}

// planDestroy collect and show units for destroying.
func (p *Project) planDestroy(opStatus *ProjectPlanningStatus) {
	for _, md := range p.OwnState.UnitsSlice() {
		if i := findMod(p.UnitsSlice(), md); i >= 0 {
			continue
		}
		diff := utils.Diff(md.GetDiffData(), nil, true)
		opStatus.Add(md, Destroy, diff, md.IsTainted())
	}
}

// planDestroyAll add all units from state for destroy.
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

func findMod(list []Unit, mod Unit) int {
	if len(list) < 1 {
		return -1
	}
	for index, m := range list {
		if mod.Key() == m.Key() {
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
					planningStatus.AddOrUpdate(unitForCheck, UpdateAsDep, colors.Fmt(colors.Yellow).Sprint("<Should be applied as a required with 'force_apply' option>"))
				}
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	return
}
