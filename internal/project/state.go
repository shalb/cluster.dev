package project

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/apex/log"
	"github.com/google/uuid"
	"github.com/shalb/cluster.dev/internal/config"
	"github.com/shalb/cluster.dev/pkg/colors"
	"github.com/shalb/cluster.dev/pkg/utils"
)

func (sp *StateProject) UpdateUnit(unit Unit) {
	// log.Warnf("UpdateUnit %v", unit.Key())
	// for u, _ := range sp.Units {
	// 	log.Warnf("     %v", u)
	// }
	sp.StateMutex.Lock()
	defer sp.StateMutex.Unlock()
	sp.Units[unit.Key()] = unit
	sp.ChangedUnits[unit.Key()] = unit
	sp.UnitLinks.Join(sp.LoaderProjectPtr.UnitLinks.ByTargetUnit(unit))
}

func (sp *StateProject) DeleteUnit(mod Unit) {
	delete(sp.Units, mod.Key())
}

func (sp *stateData) ClearULinks() {
	for linkKey, link := range sp.UnitLinks.Map() {
		if sp.Units[link.UnitKey()] == nil {
			sp.UnitLinks.Delete(linkKey)
		}
	}
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
		ProjectUUID: p.UUID,
		Units:       map[string]interface{}{},
	}
	// log.Errorf("units links: %+v\n Project: %+v", st.UnitLinks, p.UnitLinks)
	for key, unit := range p.Units {
		st.Units[key] = unit.GetState()
	}
	// Remove all unit links, that not have a target unit.
	st.ClearULinks()
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
	ProjectUUID string                 `json:"project_uuid,omitempty"`
	UnitLinks   *UnitLinksT            `json:"unit_links"`
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
	return os.WriteFile(bkFileName, loadedStateFile, 0660)
}

func (p *Project) BackupState() error {
	loadedStateFile, err := p.GetState()
	if err != nil {
		return fmt.Errorf("backup state: %w", err)
	}
	const layout = "20060102150405"
	bkFileName := filepath.Join(config.Global.WorkingDir, fmt.Sprintf("cdev.state.backup.%v", time.Now().Format(layout)))
	log.Infof("Backuping state file: %v", bkFileName)
	return os.WriteFile(bkFileName, loadedStateFile, 0660)
}

func createProjectUUID() string {
	id := uuid.New()
	return id.String()
}

func (p *Project) NewEmptyState() *StateProject {
	stateD := stateData{
		UnitLinks: &UnitLinksT{},
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
			StateMutex:       sync.Mutex{},
			InitLock:         sync.Mutex{},
			UUID:             p.UUID,
		},
		LoaderProjectPtr: p,
		ChangedUnits:     make(map[string]Unit),
	}
	return &statePrj
}

func (p *Project) LoadState() (*StateProject, error) {
	if _, err := os.Stat(config.Global.StateCacheDir); os.IsNotExist(err) {
		err := os.Mkdir(config.Global.StateCacheDir, 0755)
		if err != nil {
			return nil, fmt.Errorf("load state: create state cache dir: %w", err)
		}
	}
	err := utils.RemoveDirContent(config.Global.StateCacheDir)
	if err != nil {
		return nil, fmt.Errorf("load state: remove state cache dir: %w", err)
	}

	stateD := stateData{
		UnitLinks: &UnitLinksT{},
	}

	loadedStateFile, err := p.GetState()
	if err != nil {
		return nil, err
	}
	if len(loadedStateFile) > 0 {
		// log.Warnf("LoadState():\n%s", string(loadedStateFile))
		err = utils.JSONDecode(loadedStateFile, &stateD)
		if err != nil {
			return nil, fmt.Errorf("load state: %w", err)
		}
	}
	stateD.ClearULinks()
	p.UUID = stateD.ProjectUUID
	if p.UUID == "" {
		p.UUID = createProjectUUID()
		log.Debugf("Project UUID created: %v", p.UUID)
	} else {
		log.Debugf("Project UUID loaded from state: %v", p.UUID)
	}
	statePrj := p.NewEmptyState()
	statePrj.UnitLinks = stateD.UnitLinks
	for mName, mState := range stateD.Units {
		if mState == nil {
			continue
		}
		unit, err := NewUnitFromState(mState.(map[string]interface{}), mName, statePrj)
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
	for key := range statePrj.Units {
		log.Warnf("LoadState %v", key)
	}

	return statePrj, nil
}

func (sp *StateProject) CheckUnitChanges(unit Unit) (string, Unit) {
	unitStateCache := map[string]bool{}
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
	if unitInState.IsTainted() {
		return colors.Fmt(colors.Yellow).Sprintf("Unit is tainted!\n%v", utils.Diff(nil, unit.GetDiffData(), false)), unitInState
	}
	for _, dep := range unit.Dependencies().UniqUnits() {
		if sp.checkUnitChangesRecursive(dep, unitStateCache) {
			return colors.Fmt(colors.Yellow).Sprintf("+/- There are changes in the unit dependencies."), unitInState
		}
	}
	return "", unitInState
}

func (sp *StateProject) checkUnitChangesRecursive(unit Unit, cacheUnitChanges map[string]bool) bool {
	if unit.WasApplied() {
		return true
	}
	unitInState, exists := sp.Units[unit.Key()]
	if !exists {
		return true
	}
	unitInCache, exists := cacheUnitChanges[unit.Key()]
	if exists {
		return unitInCache
	}
	diffData := unit.GetDiffData()

	df := utils.Diff(unitInState.GetDiffData(), diffData, true)
	if len(df) > 0 {
		cacheUnitChanges[unit.Key()] = true
		return true
	}
	for dep, depUnit := range unit.Dependencies().UniqUnits() {
		_, exists := sp.ChangedUnits[dep]
		if exists {
			cacheUnitChanges[unit.Key()] = true
			return true
		}
		if sp.checkUnitChangesRecursive(depUnit, cacheUnitChanges) {
			cacheUnitChanges[unit.Key()] = true
			return true
		}
	}
	cacheUnitChanges[unit.Key()] = false
	return false
}
