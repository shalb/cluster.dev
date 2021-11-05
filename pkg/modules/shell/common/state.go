package common

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/shalb/cluster.dev/pkg/project"
	"github.com/shalb/cluster.dev/pkg/utils"
)

// UnitDiffSpec describe the pieces of StateSpec data, that will be comered in "plan" diff and should affect the unit redeployment.
type UnitDiffSpec struct {
	Outputs       map[string]string      `json:"outputs,omitempty"`
	CreateFiles   FilesListT             `json:"create_files,omitempty"`
	ApplyConf     OperationConfig        `json:"apply"`
	Env           map[string]interface{} `json:"env"`
	OutputsConfig OutputsConfigSpec      `json:"outputs_config,omitempty"`
	PreHook       *HookSpec              `json:"pre_hook,omitempty"`
	PostHook      *HookSpec              `json:"post_hook,omitempty"`
}

func (m *Unit) GetState() interface{} {
	return *m.StatePtr
}

func (m *Unit) GetUnitDiff() UnitDiffSpec {
	st := UnitDiffSpec{
		Outputs:       make(map[string]string),
		ApplyConf:     m.ApplyConf,
		CreateFiles:   m.CreateFiles,
		Env:           make(map[string]interface{}),
		OutputsConfig: m.GetOutputsConf,
	}
	if m.Env != nil {
		for key, val := range m.Env.(map[string]interface{}) {
			st.Env[key] = val
		}
	}
	for output := range m.Outputs {
		st.Outputs[output] = "<output>"
	}
	return st
}

// GetDiffData return unit representation as a data set for diff and reapply.
func (m *Unit) GetDiffData() interface{} {
	diffData := map[string]interface{}{}
	utils.JSONCopy(m.StatePtr, &diffData)
	project.ScanMarkers(&diffData, project.StateOutputsScanner, m)
	res := make(map[string]interface{})
	m.ReplaceOutputsForDiff(diffData, &res)
	return res
}

// GetStateDiffData return unit representation as a data set for diff only update state.
func (m *Unit) GetStateDiffData() interface{} {
	st := m.GetState()
	diffData := map[string]interface{}{}
	utils.JSONCopy(st, m)
	return diffData
}

func (m *Unit) LoadState(spec interface{}, modKey string, p *project.StateProject) error {

	mkSplitted := strings.Split(modKey, ".")
	if len(mkSplitted) != 2 {
		return fmt.Errorf("loading unit state: bad unit key: %v", modKey)
	}
	stackName := mkSplitted[0]
	modName := mkSplitted[1]

	err := utils.JSONCopy(spec, &m)

	if err != nil {
		return fmt.Errorf("loading unit state: can't parse state: %v", err.Error())
	}

	stack, exists := p.LoaderProjectPtr.Stack[stackName]
	if !exists {
		stack = &project.Stack{
			ProjectPtr: &p.Project,
			Name:       stackName,
		}
	}

	m.MyName = modName
	m.StackPtr = stack
	m.ProjectPtr = &p.Project
	m.SpecRaw = make(map[string]interface{})
	m.UnitMarkers = make(map[string]interface{})
	m.CacheDir = filepath.Join(m.Project().CodeCacheDir, m.Key())
	// m.Outputs = make(map[string]*project.DependencyOutput)
	err = utils.JSONCopy(m, m.StatePtr)
	if err != nil {
		return fmt.Errorf("load state: %w", err)
	}
	return nil
}

// ReplaceOutputsForDiff replace remote state markers in struct to <remote state stack.mod.output> to show in diff.
func (m *Unit) ReplaceOutputsForDiff(in, out interface{}) error {
	inJSON, err := utils.JSONEncode(in)
	if err != nil {
		return fmt.Errorf("unit diff: internal error")
	}
	inJSONstr := string(inJSON)
	depMarkers, ok := m.Project().Markers[project.OutputMarkerCatName]
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
		if strings.Contains(inJSONstr, key) {
			remoteStateRef := fmt.Sprintf("<output %s.%s.%s>", marker.StackName, marker.UnitName, marker.Output)
			replacer := strings.NewReplacer(key, remoteStateRef)
			inJSONstr = replacer.Replace(inJSONstr)
		}

	}
	return utils.JSONDecode([]byte(inJSONstr), out)
}
