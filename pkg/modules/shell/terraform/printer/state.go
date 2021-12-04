package tfmodule

import (
	"fmt"

	"github.com/shalb/cluster.dev/pkg/modules/shell/terraform/base"
	"github.com/shalb/cluster.dev/pkg/project"
	"github.com/shalb/cluster.dev/pkg/utils"
)

func (u *Unit) GetState() interface{} {
	unitState := Unit{}
	err := utils.JSONCopy(*u, &unitState)
	if err != nil {
		return fmt.Errorf("read unit '%v': create state: %w", u.Name(), err)
	}
	unitState.Unit = u.Unit.GetState().(base.Unit)
	return unitState
}

type UnitDiffSpec struct {
	base.UnitDiffSpec
	Inputs        interface{} `json:"inputs"`
	OutputsConfig *string     `json:"-"`
}

func (m *Unit) GetUnitDiff() UnitDiffSpec {
	diff := m.Unit.GetUnitDiff()
	st := UnitDiffSpec{
		UnitDiffSpec:  diff,
		Inputs:        m.Inputs,
		OutputsConfig: nil,
	}
	return st
}

func (m *Unit) GetDiffData() interface{} {
	st := m.GetUnitDiff()
	res := map[string]interface{}{}
	utils.JSONCopy(st, &res)
	project.ScanMarkers(res, base.StringRemStScanner, m)
	project.ScanMarkers(res, project.StateOutputsScanner, m)
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
	return err
}
