package tfmodule

import (
	"fmt"

	"github.com/shalb/cluster.dev/pkg/modules/terraform/common"
	"github.com/shalb/cluster.dev/pkg/project"
	"github.com/shalb/cluster.dev/pkg/utils"
)

type State struct {
	common.StateSpecCommon
	Source  string      `json:"source"`
	Version string      `json:"version,omitempty"`
	ModType string      `json:"type"`
	Inputs  interface{} `json:"inputs"`
}

type StateDiff struct {
	common.StateSpecDiffCommon
	Source  string      `json:"source"`
	Version string      `json:"version,omitempty"`
	Inputs  interface{} `json:"inputs"`
}

func (m *tfModule) GetState() interface{} {
	st := m.GetStateCommon()
	stTf := State{
		StateSpecCommon: st,
		Inputs:          m.inputs,
		ModType:         m.KindKey(),
		Source:          m.source,
		Version:         m.version,
	}
	return stTf
}

func (m *tfModule) GetDiffData() interface{} {
	st := m.GetStateDiffCommon()
	stTf := StateDiff{
		StateSpecDiffCommon: st,
		Inputs:              m.inputs,
		Source:              m.source,
		Version:             m.version,
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

func (m *tfModule) LoadState(stateData interface{}, modKey string, p *project.StateProject) error {
	s := State{}
	err := utils.JSONInterfaceToType(stateData, &s)
	if err != nil {
		return fmt.Errorf("load state: %v", err.Error())
	}
	m.inputs = s.Inputs.(map[string]interface{})
	m.source = s.Source
	m.version = s.Version
	return m.LoadStateBase(s.StateSpecCommon, modKey, p)
}
