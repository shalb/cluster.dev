package base

import (
	"fmt"

	"github.com/shalb/cluster.dev/pkg/modules/shell/common"
	"github.com/shalb/cluster.dev/pkg/project"
	"github.com/shalb/cluster.dev/pkg/utils"
)

type UnitDiffSpec struct {
	common.UnitDiffSpec
	Manifests interface{} `json:"manifests"`
}

func (u *Unit) GetState() interface{} {
	if u.SavedState != nil {
		return u.SavedState
	}
	unitState := Unit{}
	err := utils.JSONCopy(*u, &unitState)
	if err != nil {
		return fmt.Errorf("read unit '%v': create state: %w", u.Name(), err)
	}
	unitState.Unit = *u.Unit.GetStateUnit()
	unitState.ApplyConf = nil
	unitState.DestroyConf = nil
	unitState.InitConf = nil
	unitState.PlanConf = nil
	unitState.Env = nil
	unitState.OutputParsers = nil
	unitState.CreateFiles = nil
	unitState.WorkDir = ""
	return unitState
}

func (u *Unit) GetUnitDiff() UnitDiffSpec {
	diff := u.Unit.GetUnitDiff()
	st := UnitDiffSpec{
		UnitDiffSpec: diff,
		Manifests:    u.GetManifestsMap(),
	}
	st.UnitDiffSpec.ApplyConf = nil
	st.UnitDiffSpec.ApplyConf = nil
	st.UnitDiffSpec.Env = nil
	st.UnitDiffSpec.CreateFiles = nil
	st.UnitDiffSpec.OutputsConfig = nil
	return st
}

func (u *Unit) GetDiffData() interface{} {
	diff := u.GetUnitDiff()
	diffData := map[string]interface{}{}
	utils.JSONCopy(diff, &diffData)
	project.ScanMarkers(&diffData, project.StateOutputsReplacer, u)
	u.ReplaceOutputsForDiff(diffData, &diffData)
	return diffData

}

func (u *Unit) GetStateDiffData() interface{} {
	// return nothing
	return ""
}

func (u *Unit) LoadState(spec interface{}, modKey string, p *project.StateProject) error {
	// log.Fatal("BOOO")
	err := u.Unit.LoadState(spec, modKey, p)
	if err != nil {
		return err
	}
	// log.Warnf("%+v", spec)
	err = utils.JSONCopy(spec, &u)
	if err != nil {
		return fmt.Errorf("loading unit state: can't parse state: %v", err.Error())
	}
	u.fillShellUnit()
	return err
}
