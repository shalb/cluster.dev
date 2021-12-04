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
	CreateFiles   *FilesListT            `json:"create_files,omitempty"`
	ApplyConf     *OperationConfig       `json:"apply,omitempty"`
	Env           map[string]interface{} `json:"env,omitempty"`
	OutputsConfig *OutputsConfigSpec     `json:"outputs_config,omitempty"`
	PreHook       *HookSpec              `json:"pre_hook,omitempty"`
	PostHook      *HookSpec              `json:"post_hook,omitempty"`
}

func (u *Unit) GetState() interface{} {
	unitState := Unit{}
	err := utils.JSONCopy(*u, &unitState)
	if err != nil {
		return fmt.Errorf("read unit '%v': create state: %w", u.Name(), err)
	}
	u.Outputs = &project.DependenciesOutputsT{
		List: nil,
	}
	return unitState
}

func (u *Unit) GetUnitDiff() UnitDiffSpec {
	st := UnitDiffSpec{
		Outputs:       make(map[string]string),
		ApplyConf:     u.ApplyConf,
		CreateFiles:   u.CreateFiles,
		Env:           make(map[string]interface{}),
		OutputsConfig: u.GetOutputsConf,
	}
	if u.Env != nil {
		for key, val := range u.Env.(map[string]interface{}) {
			st.Env[key] = val
		}
	}
	if u.Outputs != nil {
		for output := range u.Outputs.List {
			st.Outputs[output] = "<output>"
		}
	}
	return st
}

// GetDiffData return unit representation as a data set for diff and reapply.
func (u *Unit) GetDiffData() interface{} {
	diffData := map[string]interface{}{}
	diff := u.GetUnitDiff()
	utils.JSONCopy(diff, &diffData)
	project.ScanMarkers(&diffData, project.StateOutputsScanner, u)
	u.ReplaceOutputsForDiff(diffData, &diffData)
	return diffData
}

// GetStateDiffData return unit representation as a data set for diff only update state.
func (u *Unit) GetStateDiffData() interface{} {
	st := u.GetState()
	diffData := map[string]interface{}{}
	utils.JSONCopy(u, st)
	return diffData
}

func (u *Unit) LoadState(spec interface{}, modKey string, p *project.StateProject) error {

	mkSplitted := strings.Split(modKey, ".")
	if len(mkSplitted) != 2 {
		return fmt.Errorf("loading unit state: bad unit key: %v", modKey)
	}
	stackName := mkSplitted[0]
	modName := mkSplitted[1]

	err := utils.JSONCopy(spec, &u)

	if err != nil {
		return fmt.Errorf("loading unit state: can't parse state: %v", err.Error())
	}
	stack, exists := p.LoaderProjectPtr.Stacks[stackName]
	if !exists {
		stack = &project.Stack{
			ProjectPtr: &p.Project,
			Name:       stackName,
		}
	}
	backend, exists := p.LoaderProjectPtr.Backends[u.BackendName]
	if !exists {
		return fmt.Errorf("load unit from state: backend '%v' does not exists in curent project", u.BackendName)
	}

	u.MyName = modName
	u.StackPtr = stack
	u.ProjectPtr = &p.Project
	u.SpecRaw = make(map[string]interface{})
	//u.UnitMarkers = make(map[string]interface{})
	u.CacheDir = filepath.Join(u.Project().CodeCacheDir, u.Key())
	u.BackendPtr = &backend
	return nil
}

// ReplaceOutputsForDiff replace remote state markers in struct to <remote state stack.mod.output> to show in diff.
func (u *Unit) ReplaceOutputsForDiff(in, out interface{}) error {
	inJSON, err := utils.JSONEncode(in)
	if err != nil {
		return fmt.Errorf("unit diff: internal error")
	}
	inJSONstr := string(inJSON)
	depMarkers, ok := u.Project().Markers[project.OutputMarkerCatName]
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
