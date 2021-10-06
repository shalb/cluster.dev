package common

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/shalb/cluster.dev/pkg/project"
	"github.com/shalb/cluster.dev/pkg/utils"
)

// StateSpec the unit's data to
type StateSpec struct {
	WorkDir         string                               `json:"work_dir"`
	BackendName     string                               `json:"backend_name"`
	Markers         map[string]interface{}               `json:"markers,omitempty"`
	Dependencies    []*project.DependencyOutput          `json:"dependencies,omitempty"`
	CustomStateData map[string]interface{}               `json:"custom_state_data,omitempty"`
	CreateFiles     []CreateFileRepresentation           `json:"create_files,omitempty"`
	ModType         string                               `json:"type"`
	ApplyConf       OperationConfig                      `json:"apply"`
	Env             map[string]interface{}               `json:"env"`
	Outputs         map[string]*project.DependencyOutput `json:"outputs,omitempty"`
	OutputsConfig   OutputsConfigSpec                    `json:"outputs_config,omitempty"`
}

// StateSpecDiff describe the pieces of StateSpec data, that will be comered in "plan" diff and should affect the unit redeployment.
type StateSpecDiff struct {
	Outputs         map[string]string          `json:"outputs,omitempty"`
	CustomStateData map[string]interface{}     `json:"custom_state_data,omitempty"`
	CreateFiles     []CreateFileRepresentation `json:"create_files,omitempty"`
	ApplyConf       OperationConfig            `json:"apply"`
	Env             map[string]interface{}     `json:"env"`
	OutputsConfig   OutputsConfigSpec          `json:"outputs_config,omitempty"`
}

func (m *Unit) buildState() *StateSpec {
	st := StateSpec{
		BackendName:     m.backendPtr.Name(),
		Markers:         m.markers,
		Dependencies:    m.dependencies,
		WorkDir:         m.WorkDir,
		Outputs:         m.outputs,
		CustomStateData: make(map[string]interface{}),
		ModType:         m.KindKey(),
		ApplyConf: OperationConfig{
			Commands: make([]interface{}, len(m.ApplyConf.Commands)),
		},
		Env:           make(map[string]interface{}),
		CreateFiles:   m.CreateFiles,
		OutputsConfig: m.GetOutputsConf,
	}
	for i := range m.ApplyConf.Commands {
		st.ApplyConf.Commands[i] = m.ApplyConf.Commands[i]
	}
	if m.Env != nil {
		for key, val := range m.Env.(map[string]interface{}) {
			st.Env[key] = val
		}
	}
	if len(m.dependencies) == 0 {
		st.Dependencies = make([]*project.DependencyOutput, 0)
	}
	return &st
}

func (m *Unit) GetState() interface{} {
	if m.statePtr == nil {
		m.statePtr = m.buildState()
	}
	return m.statePtr
}

func (m *Unit) GetStateDiff() StateSpecDiff {
	st := StateSpecDiff{
		Outputs:       make(map[string]string),
		ApplyConf:     m.ApplyConf,
		CreateFiles:   m.CreateFiles,
		Env:           make(map[string]interface{}),
		OutputsConfig: m.GetOutputsConf,
	}
	if m.Env != nil {
		for key, val := range m.Env.(map[string]interface{}) {
			st.Env[key] = val
		}
	}
	for output := range m.outputs {
		st.Outputs[output] = "<output>"
	}
	return st
}

func (m *Unit) GetDiffData() interface{} {
	st := m.GetStateDiff()
	diffData := map[string]interface{}{}
	utils.JSONInterfaceToType(st, &diffData)
	project.ScanMarkers(&diffData, project.StateOutputsScanner, m)
	return diffData
}

func (m *Unit) LoadState(spec interface{}, modKey string, p *project.StateProject) error {

	mkSplitted := strings.Split(modKey, ".")
	if len(mkSplitted) != 2 {
		return fmt.Errorf("loading unit state: bad unit key: %v", modKey)
	}
	stackName := mkSplitted[0]
	modName := mkSplitted[1]
	var mState StateSpec
	err := utils.JSONInterfaceToType(spec, &mState)

	if err != nil {
		return fmt.Errorf("loading unit state: can't convert state data: %v", err.Error())
	}

	backend, exists := p.LoaderProjectPtr.Backends[mState.BackendName]
	if !exists {
		return fmt.Errorf("load unit from state: backend '%v' does not exists in curent project", mState.BackendName)
	}
	stack, exists := p.LoaderProjectPtr.Stack[stackName]
	if !exists {
		stack = &project.Stack{
			ProjectPtr:  &p.Project,
			Backend:     backend,
			Name:        stackName,
			BackendName: mState.BackendName,
		}
	}
	bPtr, exists := stack.ProjectPtr.Backends[stack.BackendName]
	if !exists {
		return fmt.Errorf("Backend '%s' not found, stack: '%s'", stack.BackendName, stack.Name)
	}
	m.MyName = modName
	m.stackPtr = stack
	m.projectPtr = &p.Project
	m.dependencies = mState.Dependencies
	m.backendPtr = bPtr
	m.filesList = make(map[string][]byte)
	m.specRaw = make(map[string]interface{})
	m.markers = make(map[string]interface{})
	m.cacheDir = filepath.Join(m.ProjectPtr().CodeCacheDir, m.Key())
	m.ApplyConf = mState.ApplyConf
	m.Env = mState.Env
	m.CreateFiles = mState.CreateFiles
	m.outputs = make(map[string]*project.DependencyOutput)
	m.GetOutputsConf = mState.OutputsConfig
	m.WorkDir = mState.WorkDir

	for key := range mState.Outputs {
		m.outputs[key] = &project.DependencyOutput{
			Output: key,
		}
	}
	return nil
}
