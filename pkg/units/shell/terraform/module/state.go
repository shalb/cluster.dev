package tfmodule

import (
	"fmt"
	"strings"

	"github.com/apex/log"

	"github.com/shalb/cluster.dev/pkg/project"
	"github.com/shalb/cluster.dev/pkg/units/shell/terraform/base"
	"github.com/shalb/cluster.dev/pkg/utils"
)

type UnitDiffSpec struct {
	base.UnitDiffSpec
	Source       string              `json:"source"`
	Version      string              `json:"version,omitempty"`
	Inputs       interface{}         `json:"inputs,omitempty"`
	ModulesFiles map[string][]string `json:"module_files,omitempty"`
	// LocalModule  *common.FilesListT  `json:"local_module,omitempty"`
}

func (u *Unit) GetStateUnit() *Unit {
	unitState := Unit{}
	err := utils.JSONCopy(*u, &unitState)
	if err != nil {
		log.Fatalf("read unit '%v': create state: %w", u.Name(), err)
	}
	unitState.Unit = *u.Unit.GetStateUnit()
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
		Source:       u.Source,
		Version:      u.Version,
		Inputs:       u.Inputs,
	}
	//stt := m.GetState()
	//sttjson, _ := utils.JSONEncodeString(stt)
	// log.Warnf("Module State: %v", sttjson)
	filesListDiff := map[string][]string{}
	if u.LocalModule != nil {
		for _, file := range *u.LocalModule {
			fileLines := strings.Split(file.Content, "\n")
			if len(fileLines) < 2 {
				filesListDiff[file.FileName] = []string{file.Content}
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
	}
	st.ModulesFiles = filesListDiff
	// log.Warnf("%v", st.ModulesFiles)
	return st
}

func (u *Unit) GetDiffData() interface{} {
	st := u.GetUnitDiff()
	res := map[string]interface{}{}
	utils.JSONCopy(st, &res)
	project.ScanMarkers(res, base.StringRemStScanner, u)
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
	return nil
}
