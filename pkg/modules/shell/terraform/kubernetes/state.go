package kubernetes

import (
	"fmt"

	"github.com/shalb/cluster.dev/pkg/modules/shell/terraform/base"
	"github.com/shalb/cluster.dev/pkg/project"
	"github.com/shalb/cluster.dev/pkg/utils"
)

func (m *Unit) GetState() interface{} {
	m.StatePtr.Unit = m.Unit.GetState().(base.Unit)
	return *m.StatePtr
}

type UnitDiffSpec struct {
	base.UnitDiffSpec
	ProviderConf ProviderConfigSpec `json:"provider_conf"`
	Inputs       interface{}        `json:"inputs"`
}

func (m *Unit) GetUnitDiff() UnitDiffSpec {
	diff := m.Unit.GetUnitDiff()
	st := UnitDiffSpec{
		UnitDiffSpec: diff,
		ProviderConf: m.ProviderConf,
		Inputs:       m.Inputs,
	}
	return st
}

func (m *Unit) GetDiffData() interface{} {
	st := m.GetUnitDiff()
	res := map[string]interface{}{}
	utils.JSONCopy(st, &res)
	project.ScanMarkers(res, base.StateRemStScanner, m)
	return res
}

func (m *Unit) LoadState(stateData interface{}, modKey string, p *project.StateProject) error {
	err := m.Unit.LoadState(stateData, modKey, p)
	if err != nil {
		return err
	}
	err = utils.JSONCopy(stateData, m)
	if err != nil {
		return fmt.Errorf("load state: %v", err.Error())
	}
	m.StatePtr = &Unit{
		Unit: m.Unit,
	}
	err = utils.JSONCopy(m, m.StatePtr)
	return err
}
