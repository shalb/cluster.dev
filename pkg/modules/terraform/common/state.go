package common

import (
	"fmt"
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
	Markers          map[string]string           `json:"markers,omitempty"`
	Dependencies     []StateDep                  `json:"dependencies,omitempty"`
	RequiredProvider map[string]RequiredProvider `json:"required_providers,omitempty"`
}

type StateSpecDiffCommon struct {
	// BackendName string      `json:"backend_name"`
	PreHook   *hookSpec   `json:"pre_hook,omitempty"`
	PostHook  *hookSpec   `json:"post_hook,omitempty"`
	Providers interface{} `json:"providers,omitempty"`
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
	}
	return st
}

func (m *Module) LoadStateBase(spec StateCommon, modKey string, p *project.StateProject) error {

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
	m.name = modName
	m.infraPtr = infra
	m.projectPtr = &p.Project
	m.dependencies = modDeps
	m.backendPtr = bPtr
	m.expectedOutputs = map[string]bool{}
	m.filesList = map[string][]byte{}
	m.specRaw = map[string]interface{}{}
	m.markers = map[string]string{}
	m.preHook = mState.PreHook
	m.postHook = mState.PostHook
	m.providers = mState.Providers
	m.requiredProviders = mState.RequiredProvider
	return nil
}

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
	for key, marker := range depMarkers.(map[string]*project.Dependency) {
		if strings.Contains(inJSONstr, key) {
			remoteStateRef := fmt.Sprintf("<remoteState %s.%s.%s>", marker.InfraName, marker.ModuleName, marker.Output)
			replacer := strings.NewReplacer(key, remoteStateRef)
			inJSONstr = replacer.Replace(inJSONstr)
		}

	}
	return utils.JSONDecode([]byte(inJSONstr), out)
}
