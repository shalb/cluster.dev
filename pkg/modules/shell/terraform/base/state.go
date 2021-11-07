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
	CreateFiles   bool        `json:"-"`
	ApplyConf     bool        `json:"-"`
	Env           bool        `json:"-"`
	OutputsConfig bool        `json:"-"`
	Providers     interface{} `json:"providers,omitempty"`
}

func (m *Unit) GetState() interface{} {
	m.StatePtr.ApplyConf = nil
	m.StatePtr.DestroyConf = nil
	m.StatePtr.InitConf = nil
	m.StatePtr.PlanConf = nil
	m.StatePtr.CreateFiles = nil
	m.StatePtr.Env = nil
	m.StatePtr.OutputParsers = nil
	m.StatePtr.WorkDir = ""
	return *m.StatePtr
}

func (m *Unit) GetUnitDiff() UnitDiffSpec {
	diff := m.Unit.GetUnitDiff()
	st := UnitDiffSpec{
		UnitDiffSpec: diff,
		Providers:    m.Providers,
	}
	st.UnitDiffSpec.ApplyConf = nil
	st.UnitDiffSpec.ApplyConf = nil
	st.UnitDiffSpec.CreateFiles = nil
	st.UnitDiffSpec.Env = nil

	return st
}

func (m *Unit) GetDiffData() interface{} {
	diff := m.GetUnitDiff()
	for output := range m.Outputs {
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
	err = utils.JSONCopy(spec, &m)

	if err != nil {
		return fmt.Errorf("loading unit state: can't parse state: %v", err.Error())
	}
	err = utils.JSONCopy(m, m.StatePtr)
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
