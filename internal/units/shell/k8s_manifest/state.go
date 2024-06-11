package base

import (
	"fmt"

	"github.com/apex/log"

	"github.com/shalb/cluster.dev/internal/project"
	"github.com/shalb/cluster.dev/internal/units/shell/common"
	"github.com/shalb/cluster.dev/pkg/utils"
)

type UnitDiffSpec struct {
	common.UnitDiffSpec
	Manifests interface{} `json:"manifests"`
}

func (u *Unit) GetState() project.Unit {
	if u.SavedState != nil {
		return u.SavedState
	}
	unitState := Unit{}
	err := utils.JSONCopy(u, &unitState)
	if err != nil {
		log.Fatalf("read unit '%v': create state: %w", u.Name(), err)
		// return nil
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
	return &unitState
}

func (u *Unit) GetUnitDiff() UnitDiffSpec {
	diff := u.Unit.GetUnitDiff()
	manifests, _ := u.GetManifestsMap()
	st := UnitDiffSpec{
		UnitDiffSpec: diff,
		Manifests:    manifests,
	}
	st.UnitDiffSpec.ApplyConf = nil
	st.UnitDiffSpec.ApplyConf = nil
	st.UnitDiffSpec.Env = nil
	// st.UnitDiffSpec.CreateFiles = nil
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
