package tfmodule

import (
	"fmt"

	"github.com/apex/log"

	"github.com/shalb/cluster.dev/pkg/project"
	"github.com/shalb/cluster.dev/pkg/units/shell/terraform/base"
	"github.com/shalb/cluster.dev/pkg/utils"
)

func (u *Unit) GetStateUnit() *Unit {
	unitState := Unit{}
	err := utils.JSONCopy(*u, &unitState)
	if err != nil {
		log.Fatalf("read unit '%v': create state: %w", u.Name(), err)
	}
	unitState.OutputRaw = u.OutputRaw
	unitState.Unit = *u.Unit.GetStateUnit()
	return &unitState
}

func (u *Unit) GetState() interface{} {
	if u.StateData != nil {
		return u.StateData
	}
	return *u.GetStateUnit()
}

type UnitDiffSpec struct {
	base.UnitDiffSpec
	Inputs        interface{} `json:"inputs"`
	OutputsConfig *string     `json:"-"`
}

func (u *Unit) GetUnitDiff() UnitDiffSpec {
	diff := u.Unit.GetUnitDiff()
	st := UnitDiffSpec{
		UnitDiffSpec:  diff,
		Inputs:        u.Inputs,
		OutputsConfig: nil,
	}
	return st
}

func (u *Unit) GetDiffData() interface{} {
	st := u.GetUnitDiff()
	res := map[string]interface{}{}
	utils.JSONCopy(st, &res)
	project.ScanMarkers(res, base.StringRemStScanner, u)
	project.ScanMarkers(res, project.StateOutputsReplacer, u)
	return res
}

func (u *Unit) LoadState(stateData interface{}, modKey string, p *project.StateProject) error {
	err := u.Unit.LoadState(stateData, modKey, p)
	if err != nil {
		return err
	}
	err = utils.JSONCopy(stateData, u)
	if err != nil {
		return fmt.Errorf("load state: %v", err.Error())
	}
	return err
}
