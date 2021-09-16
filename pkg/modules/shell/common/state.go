package common

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/shalb/cluster.dev/pkg/project"
	"github.com/shalb/cluster.dev/pkg/utils"
)

type StateDep struct {
	Stack  string `json:"infra"`
	Module string `json:"module"`
}

type StateSpec struct {
	BackendName     string                     `json:"backend_name"`
	Markers         map[string]interface{}     `json:"markers,omitempty"`
	Dependencies    []StateDep                 `json:"dependencies,omitempty"`
	Outputs         map[string]bool            `json:"outputs,omitempty"`
	CustomStateData map[string]interface{}     `json:"custom_state_data"`
	CreateFiles     []CreateFileRepresentation `json:"create_files,omitempty"`
	ModType         string                     `json:"type"`
	ApplyConf       OperationConfig            `json:"apply"`
	Env             interface{}                `json:"env"`
}

type StateSpecDiff struct {
	Outputs         map[string]string          `json:"outputs,omitempty"`
	CustomStateData map[string]interface{}     `json:"custom_state_data"`
	CreateFiles     []CreateFileRepresentation `json:"create_files,omitempty"`
	ApplyConf       OperationConfig            `json:"apply"`
	Env             interface{}                `json:"env"`
}

type StateCommon interface {
}

func (m *Module) GetState() interface{} {
	deps := make([]StateDep, len(m.dependencies))
	for i, dep := range m.dependencies {
		deps[i].Stack = dep.StackName
		deps[i].Module = dep.ModuleName
	}
	st := StateSpec{
		BackendName:     m.backendPtr.Name(),
		Markers:         m.markers,
		Dependencies:    deps,
		Outputs:         make(map[string]bool),
		CustomStateData: make(map[string]interface{}),
		ModType:         m.KindKey(),
		ApplyConf:       m.ApplyConf,
		Env:             m.Env,
		CreateFiles:     m.CreateFiles,
	}
	for key := range m.expectedOutputs {
		st.Outputs[key] = true
	}
	if len(m.dependencies) == 0 {
		st.Dependencies = []StateDep{}
	}
	return st
}

func (m *Module) GetStateDiff() StateSpecDiff {
	deps := make([]StateDep, len(m.dependencies))
	for i, dep := range m.dependencies {
		deps[i].Stack = dep.StackName
		deps[i].Module = dep.ModuleName
	}
	st := StateSpecDiff{
		Outputs:         map[string]string{},
		CustomStateData: make(map[string]interface{}),
	}
	for output := range m.expectedOutputs {
		st.Outputs[output] = "<output>"
	}
	st.ApplyConf = m.ApplyConf
	st.Env = m.Env
	st.CreateFiles = m.CreateFiles

	return st
}

func (m *Module) GetDiffData() interface{} {
	st := m.GetStateDiff()
	diffData := map[string]interface{}{}
	utils.JSONInterfaceToType(st, &diffData)
	project.ScanMarkers(&diffData, project.StateOutputsScanner, m)
	return diffData
}

func (m *Module) LoadState(spec interface{}, modKey string, p *project.StateProject) error {

	mkSplitted := strings.Split(modKey, ".")
	if len(mkSplitted) != 2 {
		return fmt.Errorf("loading module state: bad module key: %v", modKey)
	}
	stackName := mkSplitted[0]
	modName := mkSplitted[1]
	var mState StateSpec
	err := utils.JSONInterfaceToType(spec, &mState)

	if err != nil {
		return fmt.Errorf("loading module state: can't convert state data: %v", err.Error())
	}

	backend, exists := p.LoaderProjectPtr.Backends[mState.BackendName]
	if !exists {
		return fmt.Errorf("load module from state: backend '%v' does not exists in curent project", mState.BackendName)
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

	modDeps := make([]*project.DependencyOutput, len(mState.Dependencies))
	for i, dep := range mState.Dependencies {
		modDeps[i] = &project.DependencyOutput{
			ModuleName: dep.Module,
			StackName:  dep.Stack,
		}
	}
	bPtr, exists := stack.ProjectPtr.Backends[stack.BackendName]
	if !exists {
		return fmt.Errorf("Backend '%s' not found, stack: '%s'", stack.BackendName, stack.Name)
	}
	m.MyName = modName
	m.stackPtr = stack
	m.projectPtr = &p.Project
	m.dependencies = modDeps
	m.backendPtr = bPtr
	m.filesList = make(map[string][]byte)
	m.specRaw = make(map[string]interface{})
	m.markers = make(map[string]interface{})
	m.WorkDir = filepath.Join(m.ProjectPtr().CodeCacheDir, m.Key())
	m.ApplyConf = mState.ApplyConf
	m.Env = mState.Env
	m.CreateFiles = mState.CreateFiles
	m.expectedOutputs = make(map[string]*project.DependencyOutput)

	for key := range mState.Outputs {
		m.expectedOutputs[key] = &project.DependencyOutput{
			Output: key,
		}
	}
	return nil
}
