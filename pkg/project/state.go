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

func (sp *StateProject) UpdateUnit(mod Unit) {
	sp.StateMutex.Lock()
	defer sp.StateMutex.Unlock()
	sp.Units[mod.Key()] = mod
	sp.ChangedUnits[mod.Key()] = mod
}

func (sp *StateProject) DeleteUnit(mod Unit) {
	delete(sp.Units, mod.Key())
}

type StateProject struct {
	Project
	LoaderProjectPtr *Project
	ChangedUnits     map[string]Unit
}

func (p *Project) SaveState() error {
	p.StateMutex.Lock()
	defer p.StateMutex.Unlock()
	st := stateData{
		Markers: p.Markers,
		Units:   map[string]interface{}{},
	}
	for key, mod := range p.Units {
		st.Units[key] = mod.GetState()
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
	// Modules map[string]interface{} `json:"modules"`
	Units map[string]interface{} `json:"units"`
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

	stateD := stateData{
		Markers: make(map[string]interface{}),
	}
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
			Units:            make(map[string]Unit),
			Markers:          stateD.Markers,
			Stack:            make(map[string]*Stack),
			Backends:         p.Backends,
			CodeCacheDir:     config.Global.StateCacheDir,
			StateBackendName: p.StateBackendName,
		},
		LoaderProjectPtr: p,
		ChangedUnits:     make(map[string]Unit),
	}

	if statePrj.Markers == nil {
		statePrj.Markers = make(map[string]interface{})
	}
	// for key, m := range p.Markers {
	// 	statePrj.Markers[key] = m
	// }
	utils.JSONCopy(p.Markers, statePrj.Markers)
	for mName, mState := range stateD.Units {
		log.Debugf("Loading unit from state: %v", mName)

		if mState == nil {
			continue
		}
		key, exists := mState.(map[string]interface{})["type"]
		if !exists {
			return nil, fmt.Errorf("loading state: internal error: bad unit type in state")
		}
		f, exists := UnitFactoriesMap[key.(string)]
		if !exists {
			return nil, fmt.Errorf("loading state: internal error: bad unit type in state")
		}
		unit, err := f.NewFromState(mState.(map[string]interface{}), mName, &statePrj)
		if err != nil {
			return nil, fmt.Errorf("loading state: error loading unit from state: %v", err.Error())
		}
		statePrj.Units[mName] = unit
		unit.UpdateProjectRuntimeData(&statePrj.Project)
	}
	err = statePrj.prepareUnits()
	if err != nil {
		return nil, err
	}
	p.ownState = &statePrj
	return &statePrj, nil
}

func (sp *StateProject) CheckUnitChanges(unit Unit) (string, Unit) {
	unitInState, exists := sp.Units[unit.Key()]
	if !exists {
		return utils.Diff(nil, unit.GetDiffData(), true), nil
	}

	diffData := unit.GetDiffData()
	stateDiffData := unitInState.GetDiffData()
	// m, _ := utils.JSONEncodeString(diffData)
	// log.Warnf("Diff data: %v", m)
	// sm, _ := utils.JSONEncodeString(stateDiffData)
	// log.Warnf("State diff data: %v", sm)
	// mr, _ := utils.JSONEncodeString(unitInState.Project().Markers)
	// log.Warnf("markers: %v", mr)
	// smr, _ := utils.JSONEncodeString(unitInState.Project().Markers)
	// log.Warnf("state markers: %v", smr)
	df := utils.Diff(stateDiffData, diffData, true)
	if len(df) > 0 {
		return df, unitInState
	}
	for _, u := range unit.RequiredUnits() {
		if sp.checkUnitChangesRecursive(u) {
			return colors.Fmt(colors.Yellow).Sprintf("+/- There are changes in the unit dependencies."), unitInState
		}
	}
	return "", unitInState
}

func (sp *StateProject) checkUnitChangesRecursive(unit Unit) bool {
	if unit.WasApplied() {
		return true
	}
	unitInState, exists := sp.Units[unit.Key()]
	if !exists {
		return true
	}
	diffData := unit.GetDiffData()

	df := utils.Diff(unitInState.GetDiffData(), diffData, true)
	if len(df) > 0 {
		return true
	}
	for _, u := range unit.RequiredUnits() {
		if _, exists := sp.ChangedUnits[u.Key()]; exists {
			return true
		}
		if sp.checkUnitChangesRecursive(u) {
			return true
		}
	}
	return false
}
