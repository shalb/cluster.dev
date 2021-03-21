package tfmodule

import (
	"fmt"

	"github.com/shalb/cluster.dev/pkg/modules/terraform/common"
	"github.com/shalb/cluster.dev/pkg/project"
	"github.com/shalb/cluster.dev/pkg/utils"
)

type State struct {
	common.StateSpecCommon
	ModType string      `json:"type"`
	Inputs  interface{} `json:"inputs"`
}

func (m *printer) GetState() interface{} {
	st := m.GetStateCommon()
	printer := State{
		StateSpecCommon: st,
		Inputs:          m.inputs,
		ModType:         m.KindKey(),
	}
	return printer
}

type StateDiff struct {
	common.StateSpecDiffCommon
	Inputs interface{} `json:"inputs"`
}

func (m *printer) GetDiffData() interface{} {
	st := m.GetStateDiffCommon()
	stTf := StateDiff{
		StateSpecDiffCommon: st,
		Inputs:              m.inputs,
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

func (m *printer) LoadState(stateData interface{}, modKey string, p *project.StateProject) error {
	s := State{}
	err := utils.JSONInterfaceToType(stateData, &s)
	if err != nil {
		return fmt.Errorf("load state: %v", err.Error())
	}
	m.inputs = s.Inputs.(map[string]interface{})
	return m.LoadStateBase(s.StateSpecCommon, modKey, p)
}
