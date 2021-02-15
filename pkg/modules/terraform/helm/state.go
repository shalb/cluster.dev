package helm

import "github.com/shalb/cluster.dev/pkg/modules/terraform/common"

type State struct {
	common.StateSpecCommon
	Source     string      `json:"source"`
	Kubeconfig string      `json:"kubeconfig"`
	Kind       string      `json:"kind"`
	HelmOpts   interface{} `json:"helm_opts,omitempty"`
	Sets       interface{} `json:"sets,omitempty"`
	Values     []byte      `json:"values,omitempty"`
}

func (m *helm) GetState() (interface{}, error) {
	st, err := m.GetStateCommon()
	stTf := State{
		StateSpecCommon: st,

		Kind:       m.KindKey(),
		Source:     m.source,
		Kubeconfig: m.kubeconfig,
		HelmOpts:   m.helmOpts,
		Sets:       m.sets,
		Values:     m.valuesFileContent,
	}
	return stTf, err
}
