package project

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"

	"github.com/apex/log"
	"github.com/shalb/cluster.dev/pkg/colors"
	"github.com/shalb/cluster.dev/pkg/config"
	"github.com/shalb/cluster.dev/pkg/utils"
)

func (sp *StateProject) UpdateModule(mod Module) {
	sp.StateMutex.Lock()
	defer sp.StateMutex.Unlock()
	sp.Modules[mod.Key()] = mod
	sp.ChangedModules[mod.Key()] = mod
}

func (sp *StateProject) DeleteModule(mod Module) {
	delete(sp.Modules, mod.Key())
}

type StateProject struct {
	Project
	LoaderProjectPtr *Project
	ChangedModules   map[string]Module
}

func (p *Project) SaveState() error {
	p.StateMutex.Lock()
	defer p.StateMutex.Unlock()
	st := stateData{
		Markers: p.Markers,
		Modules: map[string]interface{}{},
	}
	for key, mod := range p.Modules {
		st.Modules[key] = mod.GetState()
	}
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent(" ", " ")
	err := encoder.Encode(st)
	if err != nil {
		return fmt.Errorf("saving project state: %v", err.Error())
	}
	if p.StateBackendName != "" {
		sBk, ok := p.Backends[p.StateBackendName]
		if !ok {
			return fmt.Errorf("lock state: state backend '%v' does not found", p.StateBackendName)
		}
		return sBk.WriteState(buffer.String())
	}
	return ioutil.WriteFile(config.Global.StateLocalFileName, buffer.Bytes(), fs.ModePerm)
}

type stateData struct {
	Markers map[string]interface{} `json:"markers"`
	Modules map[string]interface{} `json:"modules"`
}

func (p *Project) LockState() error {
	if p.StateBackendName != "" {
		sBk, ok := p.Backends[p.StateBackendName]
		if !ok {
			return fmt.Errorf("lock state: state backend '%v' does not found", p.StateBackendName)
		}
		return sBk.LockState()
	}
	_, err := ioutil.ReadFile(config.Global.StateLocalLockFile)
	if err == nil {
		return fmt.Errorf("state is locked by another process")
	}
	err = ioutil.WriteFile(config.Global.StateLocalLockFile, []byte{}, os.ModePerm)
	return err
}

func (p *Project) UnLockState() error {
	if p.StateBackendName != "" {
		sBk, ok := p.Backends[p.StateBackendName]
		if !ok {
			return fmt.Errorf("lock state: state backend '%v' does not found", p.StateBackendName)
		}
		return sBk.UnlockState()
	}
	return os.Remove(config.Global.StateLocalLockFile)
}

func (p *Project) LoadState() (*StateProject, error) {
	if _, err := os.Stat(config.Global.StateCacheDir); os.IsNotExist(err) {
		err := os.Mkdir(config.Global.StateCacheDir, 0755)
		if err != nil {
			return nil, err
		}
	}
	err := removeDirContent(config.Global.StateCacheDir)
	if err != nil {
		return nil, err
	}

	stateD := stateData{}
	var stateStr string
	var loadedStateFile []byte
	if p.StateBackendName != "" {
		sBk, ok := p.Backends[p.StateBackendName]
		if !ok {
			return nil, fmt.Errorf("load state: state backend '%v' does not found", p.StateBackendName)
		}
		stateStr, err = sBk.ReadState()
		loadedStateFile = []byte(stateStr)
	} else {
		loadedStateFile, err = ioutil.ReadFile(config.Global.StateLocalFileName)
	}

	if err == nil {
		err = utils.JSONDecode(loadedStateFile, &stateD)
		if err != nil {
			return nil, err
		}
	}
	statePrj := StateProject{
		Project: Project{
			name:             p.Name(),
			secrets:          p.secrets,
			configData:       p.configData,
			configDataFile:   p.configDataFile,
			objects:          p.objects,
			Modules:          make(map[string]Module),
			Markers:          stateD.Markers,
			Stack:            make(map[string]*Stack),
			Backends:         p.Backends,
			CodeCacheDir:     config.Global.StateCacheDir,
			StateBackendName: p.StateBackendName,
		},
		LoaderProjectPtr: p,
		ChangedModules:   make(map[string]Module),
	}

	if statePrj.Markers == nil {
		statePrj.Markers = make(map[string]interface{})
	}
	for key, m := range p.Markers {
		statePrj.Markers[key] = m
	}

	for mName, mState := range stateD.Modules {
		log.Debugf("Loading module from state: %v", mName)

		key, exists := mState.(map[string]interface{})["type"]
		if !exists {
			return nil, fmt.Errorf("loading state: internal error: bad module type in state")
		}
		mod, err := ModuleFactoriesMap[key.(string)].NewFromState(mState.(map[string]interface{}), mName, &statePrj)
		if err != nil {
			return nil, fmt.Errorf("loading state: error loading module from state: %v", err.Error())
		}
		statePrj.Modules[mName] = mod
		mod.UpdateProjectRuntimeData(&statePrj.Project)
	}
	err = statePrj.prepareModules()
	if err != nil {
		return nil, err
	}
	return &statePrj, nil
}

func (sp *StateProject) CheckModuleChanges(module Module) (string, Module) {
	// log.Debugf("Check module: %v %+v", module.Key(), sp.Modules)
	moddInState, exists := sp.Modules[module.Key()]
	if !exists {
		return utils.Diff(nil, module.GetDiffData(), true), nil
	}
	var diffData interface{}
	if module != nil {
		diffData = module.GetDiffData()
	}
	df := utils.Diff(moddInState.GetDiffData(), diffData, true)
	if len(df) > 0 {
		return df, moddInState
	}
	for _, dep := range *module.Dependencies() {
		if sp.checkModuleChangesRecursive(dep.Module) {
			return colors.Fmt(colors.Yellow).Sprintf("+/- There are changes in the module dependencies."), moddInState
		}
	}
	return "", moddInState
}

func (sp *StateProject) checkModuleChangesRecursive(module Module) bool {
	// log.Debugf("Check module recu: %v deps: %v", module.Key(), *module.Dependencies())
	modNew, exists := sp.Modules[module.Key()]
	if !exists {
		return true
	}
	var diffData interface{}
	if module != nil {
		diffData = module.GetDiffData()
	}
	df := utils.Diff(diffData, modNew.GetDiffData(), true)
	if len(df) > 0 {
		return true
	}
	// log.Debugf("Check module recu: %v deps: %v", module.Key(), *module.Dependencies())
	for _, dep := range *module.Dependencies() {
		if _, exists := sp.ChangedModules[dep.Module.Key()]; exists {
			return true
		}
		if sp.checkModuleChangesRecursive(dep.Module) {
			return true
		}
	}
	return false
}
