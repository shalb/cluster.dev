package project

import (
	"fmt"
	"os"

	"github.com/apex/log"
	"github.com/paulrademacher/climenu"
	"github.com/shalb/cluster.dev/pkg/colors"
	"github.com/shalb/cluster.dev/pkg/config"
	"github.com/shalb/cluster.dev/pkg/utils"
)

// Build generate all terraform code for project.
func (p *Project) Build() error {
	for _, module := range p.Modules {
		if err := module.Build(); err != nil {
			return err
		}
	}

	return nil
}

// Destroy all modules.
func (p *Project) Destroy() error {
	fProject, err := p.LoadState()
	if err != nil {
		return err
	}
	grph := grapher{}
	if config.Global.IgnoreState {
		grph.Init(p, 1, true)
	} else {
		grph.Init(&fProject.Project, 1, true)
	}
	if grph.Len() < 1 {
		log.Info("Nothing to destroy, exiting")
		return nil
	}
	destSeq := grph.GetSequenceSet()
	if !config.Global.Force {
		destList := planDestroy(destSeq, nil)
		showPlanResults(nil, nil, destList, nil)
		respond := climenu.GetText("Continue?(yes/no)", "no")
		if respond != "yes" {
			log.Info("Destroying cancelled")
			return nil
		}
	}
	log.Info("Destroying...")
	for _, md := range destSeq {
		log.Infof(colors.Fmt(colors.LightWhiteBold).Sprintf("Destroying module '%v'", md.Key()))
		err := md.Build()
		if err != nil {
			log.Errorf("project destroy: destroying deleted module: %v", err.Error())
		}
		err = md.Destroy()
		if err != nil {
			return fmt.Errorf("project destroy: %v", err.Error())
		}
		fProject.DeleteModule(md)
		err = fProject.SaveState()
		if err != nil {
			return fmt.Errorf("project destroy: saving state: %v", err.Error())
		}
	}
	os.Remove(config.Global.StateFileName)
	return nil
}

// Apply all modules.
func (p *Project) Apply() error {
	if !config.Global.Force {
		p.Plan()
		respond := climenu.GetText("Continue?(yes/no)", "no")
		if respond != "yes" {
			log.Info("Cancelled")
			return nil
		}
	}
	log.Info("Applying...")
	grph := grapher{}
	grph.Init(p, config.Global.MaxParallel, false)

	fProject, err := p.LoadState()
	if err != nil {
		return err
	}

	StateDestroyGrph := grapher{}
	err = StateDestroyGrph.Init(&fProject.Project, 1, true)
	if err != nil {
		return err
	}

	for _, md := range StateDestroyGrph.GetSequenceSet() {
		_, exists := p.Modules[md.Key()]
		if exists {
			continue
		}
		err := md.Build()
		if err != nil {
			log.Errorf("project apply: destroying deleted module: %v", err.Error())
		}
		err = md.Destroy()
		if err != nil {
			return fmt.Errorf("project apply: destroying deleted module: %v", err.Error())
		}
		fProject.DeleteModule(md)
		err = fProject.SaveState()
		if err != nil {
			return fmt.Errorf("project apply: saving state: %v", err.Error())
		}
	}

	for {
		if grph.Len() == 0 {
			p.SaveState()
			return nil
		}
		md, fn, err := grph.GetNextAsync()
		if err != nil {
			log.Errorf("error in module %v, waiting for all running modules done.", md.Key())
			grph.Wait()
			return fmt.Errorf("error in module %v:\n%v", md.Key(), err.Error())
		}
		if md == nil {
			p.SaveState()
			return nil
		}

		go func(mod Module, finFunc func(error), stateP *StateProject) {
			diff := stateP.CheckModuleChanges(mod)
			var res error
			if len(diff) > 0 || config.Global.IgnoreState {
				log.Infof(colors.Fmt(colors.LightWhiteBold).Sprintf("Applying module '%v':", md.Key()))
				err := mod.Build()
				if err != nil {
					log.Errorf("project apply: module build error: %v", err.Error())
				}
				res := mod.Apply()
				if res == nil {
					stateP.UpdateModule(mod)
					err := stateP.SaveState()
					if err != nil {
						finFunc(err)
						return
					}
				}
				finFunc(res)
				return
			}
			log.Infof(colors.Fmt(colors.LightWhiteBold).Sprintf("Module '%v' has not changed. Skip applying.", md.Key()))
			finFunc(res)
		}(md, fn, fProject)
	}
}

// Plan and output result.
func (p *Project) Plan() error {
	fProject, err := p.LoadState()
	if err != nil {
		return err
	}

	CurrentGrph := grapher{}
	err = CurrentGrph.Init(p, 1, false)
	if err != nil {
		return err
	}

	StateGrph := grapher{}
	err = StateGrph.Init(&fProject.Project, 1, true)
	if err != nil {
		return err
	}
	stateModsSeq := StateGrph.GetSequenceSet()
	curModsSeq := CurrentGrph.GetSequenceSet()
	modsForApply := []string{}
	modsForUpdate := []string{}
	modsUnchanged := []string{}
	log.Infof(colors.Fmt(colors.LightWhiteBold).Sprintf("Checking modules in state"))
	modsForDestroy := planDestroy(stateModsSeq, curModsSeq)

	for _, md := range curModsSeq {
		_, exists := fProject.Modules[md.Key()]
		diff := fProject.CheckModuleChanges(md)
		log.Infof(colors.Fmt(colors.LightWhiteBold).Sprintf("Planning module '%v':", md.Key()))
		if len(diff) > 0 || config.Global.IgnoreState {
			if len(diff) == 0 {
				diff = colors.Fmt(colors.GreenBold).Sprint("Not changed.")
			}
			fmt.Printf("%v\n", diff)
			if exists {
				modsForUpdate = append(modsForUpdate, md.Key())
			} else {
				modsForApply = append(modsForApply, md.Key())
			}

			if config.Global.ShowTerraformPlan {
				allDepsDeployed := true
				for _, planModDep := range *md.Dependencies() {
					_, exists := fProject.Modules[planModDep.Module.Key()]
					if !exists {
						allDepsDeployed = false
						break
					}
				}
				if allDepsDeployed {
					err := md.Build()
					if err != nil {
						log.Errorf("terraform plan: module build error: %v", err.Error())
					}
					err = md.Plan()
					if err != nil {
						log.Errorf("Module '%v' terraform plan return an error: %v", md.Key(), err.Error())
					}
				} else {
					log.Warnf("The module '%v' has dependencies that have not yet been deployed. Can't show terraform plan.", md.Key())
				}
			}
		} else {
			modsUnchanged = append(modsUnchanged, md.Key())
			log.Infof(colors.Fmt(colors.GreenBold).Sprint("Not changed."))
		}
	}
	showPlanResults(modsForApply, modsForUpdate, modsForDestroy, modsUnchanged)
	return nil
}

// planDestroy collect and show modules for destroying.
func planDestroy(stateList, projList []Module) []string {
	modsForDestroy := []string{}
	for _, md := range stateList {
		if i := findMod(projList, md); i >= 0 {
			continue
		}
		diff := utils.Diff(md.GetDiffData(), nil, true)
		log.Info(colors.Fmt(colors.Red).Sprintf("Module '%v' will be destroyed:", md.Key()))
		fmt.Printf("%v\n", diff)
		modsForDestroy = append(modsForDestroy, md.Key())
	}
	return modsForDestroy
}

func findMod(list []Module, mod Module) int {
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
