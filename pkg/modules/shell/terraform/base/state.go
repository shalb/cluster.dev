package base

import (
	"fmt"
	"strings"

	"github.com/shalb/cluster.dev/pkg/modules/shell/common"
	"github.com/shalb/cluster.dev/pkg/project"
	"github.com/shalb/cluster.dev/pkg/utils"
)

type UnitDiffSpec struct {
	// BackendName string      `json:"backend_name"`
	common.UnitDiffSpec
	Providers interface{} `json:"providers,omitempty"`
}

func (u *Unit) GetState() interface{} {

	unitState := Unit{}
	err := utils.JSONCopy(*u, &unitState)
	if err != nil {
		return fmt.Errorf("read unit '%v': create state: %w", u.Name(), err)
	}
	unitState.Unit = u.Unit.GetState().(common.Unit)
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

func (m *Unit) GetUnitDiff() UnitDiffSpec {
	diff := m.Unit.GetUnitDiff()
	st := UnitDiffSpec{
		UnitDiffSpec: diff,
		Providers:    m.Providers,
	}
	st.UnitDiffSpec.ApplyConf = nil
	st.UnitDiffSpec.ApplyConf = nil
	st.UnitDiffSpec.Env = nil
	st.UnitDiffSpec.CreateFiles = nil
	st.UnitDiffSpec.OutputsConfig = nil
	return st
}

func (m *Unit) GetDiffData() interface{} {
	diff := m.GetUnitDiff()
	for output := range m.Outputs.List {
		diff.Outputs[output] = "<terraform output>"
	}
	diffData := map[string]interface{}{}
	utils.JSONCopy(diff, &diffData)
	return diffData
}

func (m *Unit) GetStateDiffData() interface{} {
	// return nothing
	return ""
}

func (m *Unit) LoadState(spec interface{}, modKey string, p *project.StateProject) error {

	err := m.Unit.LoadState(spec, modKey, p)
	if err != nil {
		return err
	}
	m.fillShellUnit()
	err = utils.JSONCopy(spec, &m)

	return err
}

// ReplaceRemoteStatesForDiff replace remote state markers in struct to <remote state stack.mod.output> to show in diff.
func (m *Unit) ReplaceRemoteStatesForDiff(in, out interface{}) error {
	inJSON, err := utils.JSONEncode(in)
	if err != nil {
		return fmt.Errorf("unit diff: internal error")
	}
	inJSONstr := string(inJSON)
	depMarkers, ok := m.Project().Markers[RemoteStateMarkerCatName]
	if !ok {
		return utils.JSONDecode([]byte(inJSONstr), out)
	}
	markersList, ok := depMarkers.(map[string]*project.DependencyOutput)
	if !ok {
		markersList := make(map[string]*project.DependencyOutput)
		err := utils.JSONCopy(depMarkers, &markersList)
		if err != nil {
			return fmt.Errorf("remote state scanner: read dependency: bad type")
		}
	}

	for key, marker := range markersList {
		//log.Warnf("marker replace: %v", key)
		if strings.Contains(inJSONstr, key) {
			remoteStateRef := fmt.Sprintf("<remoteState %s.%s.%s>", marker.StackName, marker.UnitName, marker.Output)
			replacer := strings.NewReplacer(key, remoteStateRef)
			inJSONstr = replacer.Replace(inJSONstr)
		}

	}
	return utils.JSONDecode([]byte(inJSONstr), out)
}
