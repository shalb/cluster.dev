package helm

import (
	"fmt"

	"github.com/shalb/cluster.dev/pkg/modules/terraform/common"
	"github.com/shalb/cluster.dev/pkg/project"
	"github.com/shalb/cluster.dev/pkg/utils"
)

type State struct {
	common.StateSpecCommon
	Source     string      `json:"source"`
	Kubeconfig string      `json:"kubeconfig"`
	ModType    string      `json:"type"`
	HelmOpts   interface{} `json:"helm_opts,omitempty"`
	Sets       interface{} `json:"sets,omitempty"`
	Values     []string    `json:"values,omitempty"`
}

func (m *Module) GetState() interface{} {
	st := m.GetStateCommon()
	stTf := State{
		StateSpecCommon: st,
		ModType:         m.KindKey(),
		Source:          m.source,
		Kubeconfig:      m.kubeconfig,
		HelmOpts:        m.helmOpts,
		Sets:            m.sets,
		Values:          m.valuesFilesList,
	}
	return stTf
}

type StateDiff struct {
	common.StateSpecDiffCommon
	Source   string      `json:"source"`
	HelmOpts interface{} `json:"helm_opts,omitempty"`
	Sets     interface{} `json:"sets,omitempty"`
	Values   []string    `json:"values,omitempty"`
}

func (m *Module) GetDiffData() interface{} {
	st := m.GetStateDiffCommon()
	stTf := StateDiff{
		StateSpecDiffCommon: st,
		Values:              m.valuesFilesList,
		Source:              m.source,
		HelmOpts:            m.helmOpts,
		Sets:                m.sets,
	}
	diffData := map[string]interface{}{}
	res := map[string]interface{}{}
	utils.JSONInterfaceToType(stTf, &diffData)
	m.ReplaceRemoteStatesForDiff(diffData, &res)
	return res
}

func (s *State) GetType() string {
	return s.ModType
}

func (m *Module) LoadState(stateData interface{}, modKey string, p *project.StateProject) error {
	s := State{}
	err := utils.JSONInterfaceToType(stateData, &s)
	if err != nil {
		return fmt.Errorf("load state: %v", err.Error())
	}
	m.helmOpts = s.HelmOpts.(map[string]interface{})
	m.sets = s.Sets.(map[string]interface{})
	m.source = s.Source
	m.kubeconfig = s.Kubeconfig
	m.valuesFilesList = s.Values
	return m.LoadStateCommon(s.StateSpecCommon, modKey, p)
}
