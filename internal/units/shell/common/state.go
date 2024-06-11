package common

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/apex/log"
	"github.com/shalb/cluster.dev/internal/project"
	"github.com/shalb/cluster.dev/pkg/utils"
)

// UnitDiffSpec describe the pieces of StateSpec data, that will be compered in "plan" diff and should affect the unit redeployment.
type UnitDiffSpec struct {
	Outputs map[string]string `json:"outputs,omitempty"`
	// CreateFiles     *FilesListT            `json:"create_files,omitempty"`
	CreateFilesDiff map[string][]string    `json:"create_files,omitempty"`
	ApplyConf       *OperationConfig       `json:"apply,omitempty"`
	Env             map[string]interface{} `json:"env,omitempty"`
	OutputsConfig   *OutputsConfigSpec     `json:"outputs_config,omitempty"`
	PreHook         *HookSpec              `json:"pre_hook,omitempty"`
	PostHook        *HookSpec              `json:"post_hook,omitempty"`
}

func (u *Unit) GetStateUnit() *Unit {
	unitState := Unit{}
	err := utils.JSONCopy(*u, &unitState)
	if err != nil {
		log.Fatalf("read unit '%v': create state: %w", u.Name(), err)
	}
	return &unitState
}

func (u *Unit) GetState() project.Unit {
	if u.SavedState != nil {
		return u.SavedState
	}
	res := u.GetStateUnit()
	return res
}

func (u *Unit) GetUnitDiff() UnitDiffSpec {
	st := UnitDiffSpec{
		Outputs:   make(map[string]string),
		ApplyConf: u.ApplyConf,
		// CreateFiles:   u.CreateFiles,
		Env:           make(map[string]interface{}),
		OutputsConfig: u.GetOutputsConf,
	}
	if u.Env != nil {
		for key, val := range u.Env {
			st.Env[key] = val
		}
	}
	filesListDiff := map[string][]string{}
	for _, file := range *u.CreateFiles {
		fileLines := strings.Split(file.Content, "\n")
		if len(fileLines) < 2 {
			filesListDiff[file.FileName][0] = file.Content
		} else {
			for _, line := range fileLines {
				//log.Warnf("filesListDiff %v", line)
				if line == "" {
					continue // Ignore empty lines
				}
				filesListDiff[file.FileName] = append(filesListDiff[file.FileName], line)
			}
		}
	}
	st.CreateFilesDiff = filesListDiff
	for _, link := range u.ProjectPtr.UnitLinks.ByTargetUnit(u).Map() {
		st.Outputs[link.OutputName] = "<output>"
	}
	return st
}

// GetDiffData return unit representation as a data set for diff and reapply.
func (u *Unit) GetDiffData() interface{} {
	diffData := map[string]interface{}{}
	diff := u.GetUnitDiff()
	utils.JSONCopy(diff, &diffData)
	project.ScanMarkers(&diffData, project.StateOutputsReplacer, u)
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
	stack.Backend = backend
	u.MyName = modName
	u.StackPtr = stack
	u.ProjectPtr = &p.Project
	u.SpecRaw = make(map[string]interface{})
	//u.UnitMarkers = make(map[string]interface{})
	u.CacheDir = filepath.Join(u.Project().CodeCacheDir, u.Key())
	u.BackendPtr = &backend
	err = u.readDeps()
	if err != nil {
		return fmt.Errorf("read state dependencies: %w", err)
	}
	return nil
}

// ReplaceOutputsForDiff replace remote state markers in struct to <remote state stack.mod.output> to show in diff.
func (u *Unit) ReplaceOutputsForDiff(in, out interface{}) error {
	inJSON, err := utils.JSONEncode(in)
	if err != nil {
		return fmt.Errorf("unit diff: internal error")
	}
	inJSONstr := string(inJSON)
	for key, marker := range u.Project().UnitLinks.ByLinkTypes(project.OutputLinkType).Map() {
		if strings.Contains(inJSONstr, key) {
			remoteStateRef := fmt.Sprintf("<output %s.%s.%s>", marker.TargetStackName, marker.TargetUnitName, marker.OutputName)
			replacer := strings.NewReplacer(key, remoteStateRef)
			inJSONstr = replacer.Replace(inJSONstr)
		}

	}
	return utils.JSONDecode([]byte(inJSONstr), out)
}
