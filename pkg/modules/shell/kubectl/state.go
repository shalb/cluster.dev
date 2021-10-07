package tfmodule

import (
	"fmt"

	"github.com/shalb/cluster.dev/pkg/modules/shell/common"
	"github.com/shalb/cluster.dev/pkg/project"
	"github.com/shalb/cluster.dev/pkg/utils"
)

type State struct {
	common.StateSpec
	ModType      string      `json:"type"`
	Inputs       interface{} `json:"inputs"`
	ModOutputRaw string      `json:"output"`
}

func (m *Unit) GetState() interface{} {
	st := m.Unit.GetState()
	printer := State{
		StateSpec:    st.(common.StateSpec),
		Inputs:       m.inputs,
		ModType:      m.KindKey(),
		ModOutputRaw: m.outputRaw,
	}
	return printer
}

type StateDiff struct {
	common.StateSpecDiff
	Inputs interface{} `json:"inputs"`
}

func (m *Unit) GetDiffData() interface{} {
	st := m.Unit.GetStateDiff()
	stTf := StateDiff{
		StateSpecDiff: st,
		Inputs:        m.inputs,
	}
	diffData := map[string]interface{}{}
	utils.JSONInterfaceToType(stTf, &diffData)
	return diffData
}

func (s *State) GetType() string {
	return s.ModType
}

func (m *Unit) LoadState(stateData interface{}, modKey string, p *project.StateProject) error {
	s := State{}
	err := utils.JSONInterfaceToType(stateData, &s)
	if err != nil {
		return fmt.Errorf("load state: %v", err.Error())
	}
	m.inputs = s.Inputs.(map[string]interface{})
	m.outputRaw = s.ModOutputRaw
	return m.Unit.LoadState(s.StateSpec, modKey, p)
}
