package project

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

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
	sp.UnitLinks.Join(sp.LoaderProjectPtr.UnitLinks.ByTargetUnit(mod))
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
		CdevVersion: config.Global.Version,
		UnitLinks:   p.UnitLinks,
		Units:       map[string]interface{}{},
	}
	// log.Errorf("units links: %+v\n Project: %+v", st.UnitLinks, p.UnitLinks)
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
	if p.StateBackendName == "" {
		return fmt.Errorf("internal error: empty project backend")
	}
	sBk, ok := p.Backends[p.StateBackendName]
	if !ok {
		return fmt.Errorf("lock state: state backend '%v' does not found", p.StateBackendName)
	}
	return sBk.WriteState(buffer.String())
}

type stateData struct {
	CdevVersion string                 `json:"version"`
	UnitLinks   UnitLinksT             `json:"unit_links"`
	Units       map[string]interface{} `json:"units"`
}

func (p *Project) LockState() error {
	if p.StateBackendName == "" {
		return fmt.Errorf("internal error: empty project backend")
	}

	sBk, ok := p.Backends[p.StateBackendName]
	if !ok {
		return fmt.Errorf("lock state: state backend '%v' does not found", p.StateBackendName)
	}
	return sBk.LockState()

}

func (p *Project) UnLockState() error {
	if p.StateBackendName == "" {
		return fmt.Errorf("internal error: empty project backend")
	}

	sBk, ok := p.Backends[p.StateBackendName]
	if !ok {
		return fmt.Errorf("lock state: state backend '%v' does not found", p.StateBackendName)
	}
	return sBk.UnlockState()
}

func (p *Project) GetState() ([]byte, error) {
	var stateStr string
	var err error
	var loadedStateFile []byte
	if p.StateBackendName == "" {
		return nil, fmt.Errorf("internal error: empty project backend")
	}
	sBk, ok := p.Backends[p.StateBackendName]
	if !ok {
		return nil, fmt.Errorf("get remote state data: state backend '%v' does not found", p.StateBackendName)
	}
	stateStr, err = sBk.ReadState()
	if err != nil {
		return nil, fmt.Errorf("get remote state data: %w", err)
	}
	loadedStateFile = []byte(stateStr)

	return loadedStateFile, nil
}

func (p *Project) PullState() error {
	loadedStateFile, err := p.GetState()
	if err != nil {
		return fmt.Errorf("pull state: %w", err)
	}
	bkFileName := filepath.Join(config.Global.WorkingDir, "cdev.state")
	log.Infof("Pulling state file: %v", bkFileName)
	return ioutil.WriteFile(bkFileName, loadedStateFile, 0660)
}

func (p *Project) BackupState() error {
	loadedStateFile, err := p.GetState()
	if err != nil {
		return fmt.Errorf("backup state: %w", err)
	}
	const layout = "20060102150405"
	bkFileName := filepath.Join(config.Global.WorkingDir, fmt.Sprintf("cdev.state.backup.%v", time.Now().Format(layout)))
	log.Infof("Backuping state file: %v", bkFileName)
	return ioutil.WriteFile(bkFileName, loadedStateFile, 0660)
}

func (p *Project) LoadState() (*StateProject, error) {
	if _, err := os.Stat(config.Global.StateCacheDir); os.IsNotExist(err) {
		err := os.Mkdir(config.Global.StateCacheDir, 0755)
		if err != nil {
			return nil, fmt.Errorf("load state: create state cache dir: %w", err)
		}
	}
	err := removeDirContent(config.Global.StateCacheDir)
	if err != nil {
		return nil, fmt.Errorf("load state: remove state cache dir: %w", err)
	}

	stateD := stateData{
		UnitLinks: UnitLinksT{},
	}

	loadedStateFile, err := p.GetState()
	if err == nil {
		err = utils.JSONDecode(loadedStateFile, &stateD)
		if err != nil {
			return nil, fmt.Errorf("load state: %w", err)
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
			UnitLinks:        stateD.UnitLinks,
			Stacks:           make(map[string]*Stack),
			Backends:         p.Backends,
			CodeCacheDir:     config.Global.StateCacheDir,
			StateBackendName: p.StateBackendName,
		},
		LoaderProjectPtr: p,
		ChangedUnits:     make(map[string]Unit),
	}
	// log.Warnf("StateProject. Unit links: %+v", statePrj.UnitLinks)
	// utils.JSONCopy(p.Markers, statePrj.Markers)
	for mName, mState := range stateD.Units {
		// log.Debugf("Loading unit from state: %v", mName)

		if mState == nil {
			continue
		}
		unit, err := NewUnitFromState(mState.(map[string]interface{}), mName, &statePrj)
		if err != nil {
			return nil, fmt.Errorf("loading unit from state: %v", err.Error())
		}
		statePrj.Units[mName] = unit
		unit.UpdateProjectRuntimeData(&statePrj.Project)
	}
	err = statePrj.prepareUnits()
	if err != nil {
		return nil, err
	}

	return &statePrj, nil
}

func (sp *StateProject) CheckUnitChanges(unit Unit) (string, Unit) {
	unitInState, exists := sp.Units[unit.Key()]
	if !exists {
		return utils.Diff(nil, unit.GetDiffData(), true), nil
	}

	diffData := unit.GetDiffData()
	stateDiffData := unitInState.GetDiffData()
	df := utils.Diff(stateDiffData, diffData, true)
	if len(df) > 0 {
		return df, unitInState
	}
	for _, dep := range unit.Dependencies().List {
		if sp.checkUnitChangesRecursive(dep.Unit) {
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
	for _, dep := range unit.Dependencies().List {
		if _, exists := sp.ChangedUnits[dep.Unit.Key()]; exists {
			return true
		}
		if sp.checkUnitChangesRecursive(dep.Unit) {
			return true
		}
	}
	return false
}
