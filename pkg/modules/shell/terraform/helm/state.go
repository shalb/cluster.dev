package helm

import (
	"fmt"

	"github.com/shalb/cluster.dev/pkg/modules/shell/terraform/base"
	"github.com/shalb/cluster.dev/pkg/project"
	"github.com/shalb/cluster.dev/pkg/utils"
)

type State struct {
	base.StateSpec
	Source     string      `json:"source"`
	Kubeconfig string      `json:"kubeconfig"`
	ModType    string      `json:"type"`
	HelmOpts   interface{} `json:"helm_opts,omitempty"`
	Sets       interface{} `json:"sets,omitempty"`
	Values     []string    `json:"values,omitempty"`
}

func (m *Unit) GetState() interface{} {
	st := m.Unit.GetState()
	stTf := State{
		StateSpec:  st,
		ModType:    m.KindKey(),
		Source:     m.source,
		Kubeconfig: m.kubeconfig,
		HelmOpts:   m.helmOpts,
		Sets:       m.sets,
		Values:     m.valuesFilesList,
	}
	return stTf
}

type StateDiff struct {
	base.StateSpecDiff
	Source   string      `json:"source"`
	HelmOpts interface{} `json:"helm_opts,omitempty"`
	Sets     interface{} `json:"sets,omitempty"`
	Values   []string    `json:"values,omitempty"`
}

func (m *Unit) GetDiffData() interface{} {
	st := m.Unit.GetStateDiff()
	stTf := StateDiff{
		StateSpecDiff: st,
		Values:        m.valuesFilesList,
		Source:        m.source,
		HelmOpts:      m.helmOpts,
		Sets:          m.sets,
	}
	diffData := map[string]interface{}{}
	res := map[string]interface{}{}
	utils.JSONCopy(stTf, &diffData)
	m.ReplaceRemoteStatesForDiff(diffData, &res)
	return res
}

func (s *State) GetType() string {
	return s.ModType
}

func (m *Unit) LoadState(stateData interface{}, modKey string, p *project.StateProject) error {
	s := State{}
	err := utils.JSONCopy(stateData, &s)
	if err != nil {
		return fmt.Errorf("load state: %v", err.Error())
	}
	m.helmOpts = s.HelmOpts.(map[string]interface{})
	m.sets = s.Sets.(map[string]interface{})
	m.source = s.Source
	m.kubeconfig = s.Kubeconfig
	m.valuesFilesList = s.Values
	return m.Unit.LoadState(s.StateSpec, modKey, p)
}
