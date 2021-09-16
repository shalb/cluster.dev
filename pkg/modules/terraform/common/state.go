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

type StateSpecCommon struct {
	BackendName      string                      `json:"backend_name"`
	PreHook          *hookSpec                   `json:"pre_hook,omitempty"`
	PostHook         *hookSpec                   `json:"post_hook,omitempty"`
	Providers        interface{}                 `json:"providers,omitempty"`
	Markers          map[string]interface{}      `json:"markers,omitempty"`
	Dependencies     []StateDep                  `json:"dependencies,omitempty"`
	RequiredProvider map[string]RequiredProvider `json:"required_providers,omitempty"`
	Outputs          map[string]bool             `json:"outputs,omitempty"`
}

type StateSpecDiffCommon struct {
	// BackendName string      `json:"backend_name"`
	PreHook   *hookSpec         `json:"pre_hook,omitempty"`
	PostHook  *hookSpec         `json:"post_hook,omitempty"`
	Providers interface{}       `json:"providers,omitempty"`
	Outputs   map[string]string `json:"outputs,omitempty"`
}

type StateCommon interface {
}

func (m *Module) GetStateCommon() StateSpecCommon {
	deps := make([]StateDep, len(m.dependencies))
	for i, dep := range m.dependencies {
		deps[i].Infra = dep.InfraName
		deps[i].Module = dep.ModuleName
	}
	st := StateSpecCommon{
		BackendName:      m.backendPtr.Name(),
		PreHook:          m.preHook,
		PostHook:         m.postHook,
		Providers:        m.providers,
		Markers:          m.markers,
		Dependencies:     deps,
		RequiredProvider: m.requiredProviders,
	}
	if len(m.dependencies) == 0 {
		st.Dependencies = []StateDep{}
	}
	st.Outputs = make(map[string]bool)
	for key := range m.expectedOutputs {
		st.Outputs[key] = true
	}
	return st
}

func (m *Module) GetStateDiffCommon() StateSpecDiffCommon {
	deps := make([]StateDep, len(m.dependencies))
	for i, dep := range m.dependencies {
		deps[i].Infra = dep.InfraName
		deps[i].Module = dep.ModuleName
	}
	st := StateSpecDiffCommon{
		//BackendName: m.backendPtr.Name(),
		PreHook:   m.preHook,
		PostHook:  m.postHook,
		Providers: m.providers,
		Outputs:   map[string]string{},
	}
	for output := range m.expectedOutputs {
		st.Outputs[output] = "<terraform output>"
	}
	return st
}

func (m *Module) LoadStateCommon(spec StateCommon, modKey string, p *project.StateProject) error {

	mkSplitted := strings.Split(modKey, ".")
	if len(mkSplitted) != 2 {
		return fmt.Errorf("loading module state common: bad module key: %v", modKey)
	}
	infraName := mkSplitted[0]
	modName := mkSplitted[1]
	mState, ok := spec.(StateSpecCommon)
	if !ok {
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

	modDeps := make([]*project.DependencyOutput, len(mState.Dependencies))
	for i, dep := range mState.Dependencies {
		modDeps[i] = &project.DependencyOutput{
			ModuleName: dep.Module,
			InfraName:  dep.Infra,
		}
	}
	bPtr, exists := infra.ProjectPtr.Backends[infra.BackendName]
	if !exists {
		return fmt.Errorf("Backend '%s' not found, infra: '%s'", infra.BackendName, infra.Name)
	}
	m.name = modName
	m.infraPtr = infra
	m.projectPtr = &p.Project
	m.dependencies = modDeps
	m.backendPtr = bPtr
	m.filesList = make(map[string][]byte)
	m.specRaw = make(map[string]interface{})
	m.markers = make(map[string]interface{})
	m.preHook = mState.PreHook
	m.postHook = mState.PostHook
	m.providers = mState.Providers
	m.requiredProviders = mState.RequiredProvider
	m.codeDir = filepath.Join(m.ProjectPtr().CodeCacheDir, m.Key())
	if m.expectedOutputs == nil {
		m.expectedOutputs = make(map[string]*project.DependencyOutput)
	}
	return nil
}

// ReplaceRemoteStatesForDiff replace remote state markers in struct to <remote state infra.mod.output> to show in diff.
func (m *Module) ReplaceRemoteStatesForDiff(in, out interface{}) error {
	inJSON, err := utils.JSONEncode(in)
	if err != nil {
		return fmt.Errorf("module diff: internal error")
	}
	inJSONstr := string(inJSON)
	depMarkers, ok := m.ProjectPtr().Markers[RemoteStateMarkerCatName]
	if !ok {
		return utils.JSONDecode([]byte(inJSONstr), out)
	}
	markersList, ok := depMarkers.(map[string]*project.DependencyOutput)
	if !ok {
		markersList := make(map[string]*project.DependencyOutput)
		err := utils.JSONInterfaceToType(depMarkers, &markersList)
		if err != nil {
			return fmt.Errorf("remote state scanner: read dependency: bad type")
		}
	}
	for key, marker := range markersList {
		if strings.Contains(inJSONstr, key) {
			remoteStateRef := fmt.Sprintf("<remoteState %s.%s.%s>", marker.InfraName, marker.ModuleName, marker.Output)
			replacer := strings.NewReplacer(key, remoteStateRef)
			inJSONstr = replacer.Replace(inJSONstr)
		}

	}
	return utils.JSONDecode([]byte(inJSONstr), out)
}
