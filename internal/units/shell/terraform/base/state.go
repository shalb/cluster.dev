package base

import (
	"fmt"
	"strings"

	"github.com/apex/log"
	"github.com/shalb/cluster.dev/internal/project"
	"github.com/shalb/cluster.dev/internal/units/shell/common"
	"github.com/shalb/cluster.dev/pkg/utils"
)

type UnitDiffSpec struct {
	// BackendName string      `json:"backend_name"`
	common.UnitDiffSpec
	Providers interface{} `json:"providers,omitempty"`
}

func (u *Unit) GetStateUnit() *Unit {

	unitState := Unit{
		Unit: *u.Unit.GetStateUnit(),
	}
	err := utils.JSONCopy(*u, &unitState)
	if err != nil {
		log.Fatalf("read unit '%v': internal error, create state: %w", u.Name(), err)
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
	// log.Errorf("GetState '%v' terraform %+v", unitState.Name(), unitState.Unit.Outputs)
	return &unitState
}

func (u *Unit) GetState() project.Unit {

	if u.SavedState != nil {
		return u.SavedState
	}
	return u.GetStateUnit()
}

func (u *Unit) GetUnitDiff() UnitDiffSpec {
	diff := u.Unit.GetUnitDiff()
	st := UnitDiffSpec{
		UnitDiffSpec: diff,
		Providers:    u.Providers,
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
	for _, output := range u.Project().UnitLinks.ByTargetUnit(u).ByLinkTypes(RemoteStateLinkType).Map() {
		diff.Outputs[output.OutputName] = "<terraform output>"
	}
	diffData := map[string]interface{}{}
	utils.JSONCopy(diff, &diffData)
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
	u.fillShellUnit()
	err = utils.JSONCopy(spec, &u)

	return err
}

// ReplaceRemoteStatesForDiff replace remote state markers in struct to <remote state stack.mod.output> to show in diff.
func (u *Unit) ReplaceRemoteStatesForDiff(in, out interface{}) error {
	inJSON, err := utils.JSONEncode(in)
	if err != nil {
		return fmt.Errorf("unit diff: internal error")
	}
	inJSONstr := string(inJSON)

	markersList := u.Project().UnitLinks.ByLinkTypes(RemoteStateLinkType).Map()

	if len(markersList) == 0 {
		return utils.JSONDecode([]byte(inJSONstr), out)
	}
	for key, marker := range markersList {
		//log.Warnf("marker replace: %v", key)
		if strings.Contains(inJSONstr, key) {
			remoteStateRef := fmt.Sprintf("<remoteState %s.%s.%s>", marker.TargetStackName, marker.TargetUnitName, marker.OutputName)
			replacer := strings.NewReplacer(key, remoteStateRef)
			inJSONstr = replacer.Replace(inJSONstr)
		}

	}
	return utils.JSONDecode([]byte(inJSONstr), out)
}
