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

	fProject, err := p.LoadState()
	if err != nil {
		return fmt.Errorf("project destroy: %w", err)
	}
	graph := grapher{}
	if config.Global.IgnoreState {
		graph.Init(p, 1, true)
	} else {
		graph.Init(&fProject.Project, 1, true)
	}
	defer graph.Close()
	if graph.Len() < 1 {
		log.Info("Nothing to destroy, exiting")
		return nil
	}
	destSeq := graph.GetSequenceSet()
	if !config.Global.Force {
		destList := planDestroy(destSeq, nil)
		showPlanResults(nil, nil, destList, nil)
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
	for _, md := range destSeq {
		log.Infof(colors.Fmt(colors.LightWhiteBold).Sprintf("Destroying unit '%v'", md.Key()))
		err = md.Build()
		if err != nil {
			return fmt.Errorf("project destroy: destroying deleted unit: %w", err)
		}
		err = md.Destroy()
		if err != nil {
			return fmt.Errorf("project destroy: %w", err)
		}
		fProject.DeleteUnit(md)
		err = fProject.SaveState()
		if err != nil {
			return fmt.Errorf("project destroy: saving state: %w", err)
		}
	}
	return nil
}

// Apply all units.
func (p *Project) Apply() error {
	grDebug := grapher{}
	err := grDebug.Init(p, config.Global.MaxParallel, false)
	if err != nil {
		return err
	}
	if !config.Global.Force {
		hasChanges, err := p.Plan()
		if err != nil {
			return err
		}
		if !hasChanges {
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
	gr := grapher{}
	err = gr.Init(p, config.Global.MaxParallel, false)
	if err != nil {
		return err
	}
	defer gr.Close()
	fProject, err := p.LoadState()
	if err != nil {
		return err
	}

	StateDestroyGraph := grapher{}
	err = StateDestroyGraph.Init(&fProject.Project, 1, true)
	if err != nil {
		return err
	}
	defer StateDestroyGraph.Close()
	for _, md := range StateDestroyGraph.GetSequenceSet() {
		_, exists := p.Units[md.Key()]
		if exists {
			continue
		}
		err = md.Build()
		if err != nil {
			log.Errorf("project apply: destroying deleted unit: %v", err.Error())
		}
		err = md.Destroy()
		if err != nil {
			return fmt.Errorf("project apply: destroying deleted unit: %v", err.Error())
		}
		fProject.DeleteUnit(md)
		err = fProject.SaveState()
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

		go func(mod Unit, finFunc func(error), stateP *StateProject) {
			diff, stateUnit := stateP.CheckUnitChanges(mod)
			var res error
			if len(diff) > 0 || config.Global.IgnoreState {
				log.Infof(colors.Fmt(colors.LightWhiteBold).Sprintf("Applying unit '%v':", md.Key()))
        log.Debugf("Unit %v diff: \n%v", md.Key(), diff)
				err = mod.Build()
				if err != nil {
					log.Errorf("project apply: unit build error: %v", err.Error())
					finFunc(err)
					return
				}
				res := mod.Apply()
				if res == nil {
					stateP.UpdateUnit(mod)
					err := stateP.SaveState()
					if err != nil {
						finFunc(err)
						return
					}
					err = mod.UpdateProjectRuntimeData(p)
					if err != nil {
						finFunc(err)
						return
					}
				}
				finFunc(res)
				return
			} else {
				// Copy unit from state to project (to save raw output data for some units)
				p.Units[mod.Key()] = stateUnit
			}
			log.Infof(colors.Fmt(colors.LightWhiteBold).Sprintf("Unit '%v' has not changed. Skip applying.", md.Key()))
			finFunc(res)
		}(md, fn, fProject)
	}
}

// Plan and output result.
func (p *Project) Plan() (hasChanges bool, err error) {
	fProject, err := p.LoadState()
	if err != nil {
		return
	}

	CurrentGraph := grapher{}
	err = CurrentGraph.Init(p, 1, false)
	if err != nil {
		return
	}
	defer CurrentGraph.Close()
	StateGraph := grapher{}
	err = StateGraph.Init(&fProject.Project, 1, true)
	if err != nil {
		return
	}
	defer StateGraph.Close()
	stateModsSeq := StateGraph.GetSequenceSet()
	curModsSeq := CurrentGraph.GetSequenceSet()
	modsForApply := []string{}
	modsForUpdate := []string{}
	modsUnchanged := []string{}
	changedUnits := map[string]Unit{}
	log.Infof(colors.Fmt(colors.LightWhiteBold).Sprintf("Checking units in state"))
	modsForDestroy := planDestroy(stateModsSeq, curModsSeq)

	for _, md := range curModsSeq {
		_, exists := fProject.Units[md.Key()]

		diff, stateUnit := fProject.CheckUnitChanges(md)
		log.Infof(colors.Fmt(colors.LightWhiteBold).Sprintf("Planning unit '%v':", md.Key()))
		if len(diff) > 0 || config.Global.IgnoreState {
			changedUnits[md.Key()] = md
			if len(diff) > 0 {
				fmt.Printf("%v\n", diff)
			} else {
				fmt.Println("<No changes>")
			}
			if exists {
				modsForUpdate = append(modsForUpdate, md.Key())
			} else {
				modsForApply = append(modsForApply, md.Key())
			}
			if config.Global.ShowTerraformPlan {
				err = p.ClearCacheDir()
				if err != nil {
					return false, fmt.Errorf("project plan: clear cache dir: %v", err.Error())
				}
				allDepsDeployed := true
				for _, planModDep := range md.Dependencies().Slice() {
					_, exists := fProject.Units[planModDep.Unit.Key()]
					if !exists {
						allDepsDeployed = false
						break
					}
				}
				if allDepsDeployed {
					err = md.Build()
					if err != nil {
						log.Errorf("terraform plan: unit build error: %v", err.Error())
						return
					}
					err = md.Plan()
					if err != nil {
						log.Errorf("unit '%v' terraform plan return an error: %v", md.Key(), err.Error())
						return
					}
				} else {
					log.Warnf("The unit '%v' has dependencies that have not yet been deployed. Can't show terraform plan.", md.Key())
				}
			}
		} else {
			if stateUnit != nil {
				stateUnit.UpdateProjectRuntimeData(p)
			}
			modsUnchanged = append(modsUnchanged, md.Key())
			// log.Warnf("Plan before: %+v", p.UnitLinks.List)
			// Unit was not changed. Copy unit outputs from state.
			p.UnitLinks.JoinWithDataReplace(fProject.UnitLinks.ByTargetUnit(md))
			// log.Warnf("Plan after: %+v", p.UnitLinks.List)
			log.Infof(colors.Fmt(colors.GreenBold).Sprint("Not changed."))
		}
	}
	// for k, l := range p.UnitLinks.Map() {
	// 	log.Warnf("Link: %v\n    %v - %v", k, l.OutputName, l.OutputData)
	// }
	// Check "force_apply" units.
	for _, un := range changedUnits {
		for _, dep := range un.Dependencies().List {
			if dep.Unit.ForceApply() {
				if _, exists := changedUnits[dep.Unit.Key()]; !exists {
					modsForUpdate = append(modsForUpdate, dep.Unit.Key())
					changedUnits[dep.Unit.Key()] = dep.Unit
				}
			}
		}
	}
	showPlanResults(modsForApply, modsForUpdate, modsForDestroy, modsUnchanged)
	hasChanges = len(modsForApply)+len(modsForUpdate)+len(modsForDestroy) != 0
	return
}

// planDestroy collect and show units for destroying.
func planDestroy(stateList, projList []Unit) []string {
	modsForDestroy := []string{}
	for _, md := range stateList {
		if i := findMod(projList, md); i >= 0 {
			continue
		}
		diff := utils.Diff(md.GetDiffData(), nil, true)
		log.Info(colors.Fmt(colors.Red).Sprintf("unit '%v' will be destroyed:", md.Key()))
		fmt.Printf("%v\n", diff)
		modsForDestroy = append(modsForDestroy, md.Key())
	}
	return modsForDestroy
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
