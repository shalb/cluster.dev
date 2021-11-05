package tfmodule

import (
	"fmt"

	"github.com/shalb/cluster.dev/pkg/modules/shell/terraform/base"
	"github.com/shalb/cluster.dev/pkg/project"
	"github.com/shalb/cluster.dev/pkg/utils"
)

type State struct {
	base.StateSpec
	ModType      string      `json:"type"`
	Inputs       interface{} `json:"inputs"`
	ModOutputRaw string      `json:"output"`
}

func (m *Unit) GetState() interface{} {
	st := m.Unit.GetState()
	printer := State{
		StateSpec:    st,
		Inputs:       m.inputs,
		ModType:      m.KindKey(),
		ModOutputRaw: m.outputRaw,
	}
	return printer
}

type StateDiff struct {
	base.StateSpecDiff
	Inputs interface{} `json:"inputs"`
}

func (m *Unit) GetDiffData() interface{} {
	st := m.GetStateDiff()
	stTf := StateDiff{
		StateSpecDiff: st,
		Inputs:        m.inputs,
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
	m.inputs = s.Inputs.(map[string]interface{})
	m.outputRaw = s.ModOutputRaw
	return m.Unit.LoadState(s.StateSpec, modKey, p)
}
