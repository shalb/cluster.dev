package tfmodule

import (
	"fmt"

	"github.com/shalb/cluster.dev/pkg/modules/shell/common"
	"github.com/shalb/cluster.dev/pkg/modules/shell/terraform/base"
	"github.com/shalb/cluster.dev/pkg/project"
	"github.com/shalb/cluster.dev/pkg/utils"
)

type UnitDiffSpec struct {
	base.UnitDiffSpec
	Source      string             `json:"source"`
	Version     string             `json:"version,omitempty"`
	Inputs      interface{}        `json:"inputs,omitempty"`
	LocalModule *common.FilesListT `json:"local_module,omitempty"`
}

func (u *UnitTfModule) GetState() interface{} {
	u.StatePtr.Unit = u.Unit.GetState().(base.Unit)
	return *u.StatePtr
}

func (u *UnitTfModule) GetUnitDiff() UnitDiffSpec {
	diff := u.Unit.GetUnitDiff()
	st := UnitDiffSpec{
		UnitDiffSpec: diff,
		Source:       u.Source,
		Version:      u.Version,
		Inputs:       u.Inputs,
		LocalModule:  u.LocalModule,
	}
	//stt := m.GetState()
	//sttjson, _ := utils.JSONEncodeString(stt)
	// log.Warnf("Module State: %v", sttjson)
	return st
}

func (u *UnitTfModule) GetDiffData() interface{} {
	st := u.GetUnitDiff()
	res := map[string]interface{}{}
	utils.JSONCopy(st, &res)
	project.ScanMarkers(res, base.StringRemStScanner, u)
	return res
}

func (u *UnitTfModule) LoadState(stateData interface{}, modKey string, p *project.StateProject) error {
	err := u.Unit.LoadState(stateData, modKey, p)
	if err != nil {
		return err
	}
	err = utils.JSONCopy(stateData, u)
	if err != nil {
		return fmt.Errorf("load state: %v", err.Error())
	}
	u.StatePtr = &UnitTfModule{
		Unit: u.Unit,
	}
	err = utils.JSONCopy(u, u.StatePtr)
	return err
}