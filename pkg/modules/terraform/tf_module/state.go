package tfmodule

import "github.com/shalb/cluster.dev/pkg/modules/terraform/common"

type State struct {
	common.StateSpecCommon
	Source  string      `json:"source"`
	Version string      `json:"version,omitempty"`
	Kind    string      `json:"kind"`
	Inputs  interface{} `json:"inputs"`
}

func (m *tfModule) GetState() (interface{}, error) {
	st, err := m.GetStateCommon()
	stTf := State{
		StateSpecCommon: st,
		Inputs:          m.inputs,
		Kind:            m.ModKindKey(),
		Source:          m.source,
		Version:         m.version,
	}
	return stTf, err
}
