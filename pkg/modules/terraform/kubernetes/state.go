package kubernetes

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
	Inputs     interface{} `json:"inputs"`
}

func (m *kubernetes) GetState() interface{} {
	st := m.GetStateCommon()
	stTf := State{
		StateSpecCommon: st,
		Inputs:          m.inputs,
		ModType:         m.KindKey(),
		Source:          m.source,
		Kubeconfig:      m.kubeconfig,
	}
	return stTf
}

type StateDiff struct {
	common.StateSpecDiffCommon
	Inputs interface{} `json:"inputs"`
}

func (m *kubernetes) GetDiffData() interface{} {
	st := m.GetStateDiffCommon()
	stTf := StateDiff{
		StateSpecDiffCommon: st,
		Inputs:              m.inputs,
	}
	diffData := map[string]interface{}{}
	res := map[string]interface{}{}
	utils.JSONInteffaceToType(stTf, &diffData)
	m.ReplaceRemoteStatesForDiff(diffData, &res)
	return res
}

func (s *State) GetType() string {
	return s.ModType

}

func (m *kubernetes) LoadState(stateData interface{}, modKey string, p *project.Project) error {
	s := State{}
	err := utils.JSONInteffaceToType(stateData, &s)
	if err != nil {
		return fmt.Errorf("load state: %v", err.Error())
	}
	m.inputs = s.Inputs.(map[string]interface{})
	m.source = s.Source
	m.kubeconfig = s.Kubeconfig
	return m.LoadStateBase(s.StateSpecCommon, modKey, p)
}
