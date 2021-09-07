package common

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/shalb/cluster.dev/pkg/project"
	"github.com/shalb/cluster.dev/pkg/utils"
)

type StateDep struct {
	Infra  string `json:"infra"`
	Module string `json:"module"`
}

type StateSpec struct {
	BackendName     string                 `json:"backend_name"`
	Markers         map[string]interface{} `json:"markers,omitempty"`
	Dependencies    []StateDep             `json:"dependencies,omitempty"`
	Outputs         map[string]bool        `json:"outputs,omitempty"`
	CustomStateData map[string]interface{} `json:"custom_state_data"`
	ModType         string                 `json:"type"`
}

type StateSpecDiff struct {
	Outputs         map[string]string      `json:"outputs,omitempty"`
	CustomStateData map[string]interface{} `json:"custom_state_data"`
}

type StateCommon interface {
}

func (m *Module) GetState() interface{} {
	deps := make([]StateDep, len(m.dependencies))
	for i, dep := range m.dependencies {
		deps[i].Infra = dep.InfraName
		deps[i].Module = dep.ModuleName
	}
	st := StateSpec{
		BackendName:     m.backendPtr.Name(),
		Markers:         m.markers,
		Dependencies:    deps,
		Outputs:         m.expectedOutputs,
		CustomStateData: make(map[string]interface{}),
		ModType:         m.KindKey(),
	}
	st.CustomStateData["apply_conf"] = m.ApplyConf.Commands
	st.CustomStateData["plan_conf"] = m.PlanConf.Commands
	st.CustomStateData["destroy_conf"] = m.DestroyConf.Commands
	if len(m.dependencies) == 0 {
		st.Dependencies = []StateDep{}
	}
	return st
}

func (m *Module) GetStateDiff() StateSpecDiff {
	deps := make([]StateDep, len(m.dependencies))
	for i, dep := range m.dependencies {
		deps[i].Infra = dep.InfraName
		deps[i].Module = dep.ModuleName
	}
	st := StateSpecDiff{
		Outputs:         map[string]string{},
		CustomStateData: make(map[string]interface{}),
	}
	for output := range m.expectedOutputs {
		st.Outputs[output] = "<output>"
	}
	st.CustomStateData["apply_conf"] = m.ApplyConf.Commands
	st.CustomStateData["plan_conf"] = m.PlanConf.Commands
	st.CustomStateData["destroy_conf"] = m.DestroyConf.Commands
	return st
}

func (m *Module) GetDiffData() interface{} {
	st := m.GetStateDiff()
	diffData := map[string]interface{}{}
	utils.JSONInterfaceToType(st, &diffData)
	return diffData
}

func (m *Module) LoadState(spec interface{}, modKey string, p *project.StateProject) error {

	mkSplitted := strings.Split(modKey, ".")
	if len(mkSplitted) != 2 {
		return fmt.Errorf("loading module state common: bad module key: %v", modKey)
	}
	infraName := mkSplitted[0]
	modName := mkSplitted[1]
	var mState StateSpec
	err := utils.JSONInterfaceToType(spec, &mState)

	if err != nil {
		return fmt.Errorf("loading module state common: can't convert state data, internal error")
	}

	backend, exists := p.LoaderProjectPtr.Backends[mState.BackendName]
	if !exists {
		return fmt.Errorf("load module from state: backend '%v' does not exists in curent project", mState.BackendName)
	}
	infra, exists := p.LoaderProjectPtr.Infrastructures[infraName]
	if !exists {
		infra = &project.Infrastructure{
			ProjectPtr:  &p.Project,
			Backend:     backend,
			Name:        infraName,
			BackendName: mState.BackendName,
		}
	}

	modDeps := make([]*project.Dependency, len(mState.Dependencies))
	for i, dep := range mState.Dependencies {
		modDeps[i] = &project.Dependency{
			ModuleName: dep.Module,
			InfraName:  dep.Infra,
		}
	}
	bPtr, exists := infra.ProjectPtr.Backends[infra.BackendName]
	if !exists {
		return fmt.Errorf("Backend '%s' not found, infra: '%s'", infra.BackendName, infra.Name)
	}
	m.MyName = modName
	m.infraPtr = infra
	m.projectPtr = &p.Project
	m.dependencies = modDeps
	m.backendPtr = bPtr
	m.filesList = make(map[string][]byte)
	m.specRaw = make(map[string]interface{})
	m.markers = make(map[string]interface{})
	m.WorkDir = filepath.Join(m.ProjectPtr().CodeCacheDir, m.Key())
	m.expectedOutputs = mState.Outputs
	m.ApplyConf = OperationConfig{}
	err = utils.JSONInterfaceToType(mState.CustomStateData["apply_conf"], &m.ApplyConf.Commands)
	if err != nil {
		return fmt.Errorf("load module from state: %v", err.Error())
	}
	m.PlanConf = OperationConfig{}
	err = utils.JSONInterfaceToType(mState.CustomStateData["plan_conf"], &m.PlanConf.Commands)
	if err != nil {
		return fmt.Errorf("load module from state: %v", err.Error())
	}
	if m.expectedOutputs == nil {
		m.expectedOutputs = make(map[string]bool)
	}
	return nil
}
