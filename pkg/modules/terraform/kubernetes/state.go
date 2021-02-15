package kubernetes

import "github.com/shalb/cluster.dev/pkg/modules/terraform/common"

type State struct {
	common.StateSpecCommon
	Source     string      `json:"source"`
	Kubeconfig string      `json:"kubeconfig"`
	Kind       string      `json:"kind"`
	Inputs     interface{} `json:"inputs"`
}

func (m *kubernetes) GetState() (interface{}, error) {
	st, err := m.GetStateCommon()
	stTf := State{
		StateSpecCommon: st,
		Inputs:          m.inputs,
		Kind:            m.KindKey(),
		Source:          m.source,
		Kubeconfig:      m.kubeconfig,
	}
	return stTf, err
}
