package tfmodule

import "github.com/shalb/cluster.dev/pkg/modules/terraform/common"

type State struct {
	common.StateSpecCommon
	Kind   string      `json:"kind"`
	Inputs interface{} `json:"inputs"`
}

func (m *printer) GetState() (interface{}, error) {
	st, err := m.GetStateCommon()
	printer := State{
		StateSpecCommon: st,
		Inputs:          m.inputs,
		Kind:            m.KindKey(),
	}
	return printer, err
}
