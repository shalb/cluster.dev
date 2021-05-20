package tfmodule

import (
	"fmt"

	"github.com/shalb/cluster.dev/pkg/modules/terraform/common"
	"github.com/shalb/cluster.dev/pkg/project"
	"github.com/shalb/cluster.dev/pkg/utils"
)

type State struct {
	common.StateSpecCommon
	ModType      string      `json:"type"`
	Inputs       interface{} `json:"inputs"`
	ModOutputRaw string      `json:"output"`
}

func (m *Module) GetState() interface{} {
	st := m.GetStateCommon()
	printer := State{
		StateSpecCommon: st,
		Inputs:          m.inputs,
		ModType:         m.KindKey(),
		ModOutputRaw:    m.outputRaw,
	}
	return printer
}

type StateDiff struct {
	common.StateSpecDiffCommon
	Inputs interface{} `json:"inputs"`
}

func (m *Module) GetDiffData() interface{} {
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

func (m *Module) LoadState(stateData interface{}, modKey string, p *project.StateProject) error {
	s := State{}
	err := utils.JSONInterfaceToType(stateData, &s)
	if err != nil {
		return fmt.Errorf("load state: %v", err.Error())
	}
	m.inputs = s.Inputs.(map[string]interface{})
	m.outputRaw = s.ModOutputRaw
	return m.LoadStateCommon(s.StateSpecCommon, modKey, p)
}
