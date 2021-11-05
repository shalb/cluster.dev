package base

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/shalb/cluster.dev/pkg/modules/shell/common"
	"github.com/shalb/cluster.dev/pkg/project"
	"github.com/shalb/cluster.dev/pkg/utils"
)

type StateDep struct {
	Stack string `json:"infra"`
	Unit  string `json:"unit"`
}

type StateSpec struct {
	common.Unit
	BackendName      string                      `json:"backend_name"`
	Providers        interface{}                 `json:"providers,omitempty"`
	Markers          map[string]interface{}      `json:"markers,omitempty"`
	Dependencies     []StateDep                  `json:"dependencies,omitempty"`
	RequiredProvider map[string]RequiredProvider `json:"required_providers,omitempty"`
	Outputs          map[string]bool             `json:"outputs,omitempty"`
}

type StateSpecDiff struct {
	BackendName string            `json:"backend_name"`
	Providers   interface{}       `json:"providers,omitempty"`
	Outputs     map[string]string `json:"outputs,omitempty"`
}

type StateCommon interface {
}

func (m *Unit) GetState() StateSpec {

	deps := make([]StateDep, len(m.DependenciesList))
	for i, dep := range m.DependenciesList {
		deps[i].Stack = dep.StackName
		deps[i].Unit = dep.UnitName
	}
	st := StateSpec{
		Unit:             m.Unit.GetState().(common.Unit),
		BackendName:      m.Backend().Name(),
		Providers:        m.Providers,
		RequiredProvider: m.RequiredProviders,
	}
	if len(m.DependenciesList) == 0 {
		st.Dependencies = []StateDep{}
	}
	st.Outputs = make(map[string]bool)
	for key := range m.Outputs {
		st.Outputs[key] = true
	}
	return st
}

func (m *Unit) GetStateDiff() StateSpecDiff {
	deps := make([]StateDep, len(m.DependenciesList))
	for i, dep := range m.DependenciesList {
		deps[i].Stack = dep.StackName
		deps[i].Unit = dep.UnitName
	}
	st := StateSpecDiff{
		BackendName: m.BackendPtr.Name(),
		Providers:   m.Providers,
		Outputs:     map[string]string{},
	}
	for output := range m.Outputs {
		st.Outputs[output] = "<terraform output>"
	}
	return st
}

func (m *Unit) LoadState(spec StateCommon, modKey string, p *project.StateProject) error {

	mkSplitted := strings.Split(modKey, ".")
	if len(mkSplitted) != 2 {
		return fmt.Errorf("loading unit state common: bad unit key: %v", modKey)
	}
	stackName := mkSplitted[0]
	modName := mkSplitted[1]
	mState, ok := spec.(StateSpec)
	if !ok {
		return fmt.Errorf("loading unit state common: can't convert state data, internal error")
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

	modDeps := make([]*project.DependencyOutput, len(mState.Dependencies))
	for i, dep := range mState.Dependencies {
		modDeps[i] = &project.DependencyOutput{
			UnitName:  dep.Unit,
			StackName: dep.Stack,
		}
	}
	bPtr, exists := stack.ProjectPtr.Backends[stack.BackendName]
	if !exists {
		return fmt.Errorf("Backend '%s' not found, stack: '%s'", stack.BackendName, stack.Name)
	}
	m.MyName = modName
	m.StackPtr = stack
	m.ProjectPtr = &p.Project
	m.DependenciesList = modDeps
	m.BackendPtr = bPtr
	m.SpecRaw = make(map[string]interface{})
	m.UnitMarkers = make(map[string]interface{})
	m.Providers = mState.Providers
	m.RequiredProviders = mState.RequiredProvider
	m.CacheDir = filepath.Join(m.Project().CodeCacheDir, m.Key())
	if m.Outputs == nil {
		m.Outputs = make(map[string]*project.DependencyOutput)
	}
	return nil
}

// ReplaceRemoteStatesForDiff replace remote state markers in struct to <remote state stack.mod.output> to show in diff.
func (m *Unit) ReplaceRemoteStatesForDiff(in, out interface{}) error {
	inJSON, err := utils.JSONEncode(in)
	if err != nil {
		return fmt.Errorf("unit diff: internal error")
	}
	inJSONstr := string(inJSON)
	depMarkers, ok := m.Project().Markers[RemoteStateMarkerCatName]
	if !ok {
		return utils.JSONDecode([]byte(inJSONstr), out)
	}
	markersList, ok := depMarkers.(map[string]*project.DependencyOutput)
	if !ok {
		markersList := make(map[string]*project.DependencyOutput)
		err := utils.JSONCopy(depMarkers, &markersList)
		if err != nil {
			return fmt.Errorf("remote state scanner: read dependency: bad type")
		}
	}
	for key, marker := range markersList {
		if strings.Contains(inJSONstr, key) {
			remoteStateRef := fmt.Sprintf("<remoteState %s.%s.%s>", marker.StackName, marker.UnitName, marker.Output)
			replacer := strings.NewReplacer(key, remoteStateRef)
			inJSONstr = replacer.Replace(inJSONstr)
		}

	}
	return utils.JSONDecode([]byte(inJSONstr), out)
}
