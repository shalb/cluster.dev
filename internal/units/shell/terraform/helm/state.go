package helm

import (
	"fmt"
	"strings"

	"github.com/apex/log"

	"github.com/shalb/cluster.dev/internal/project"
	"github.com/shalb/cluster.dev/internal/units/shell/terraform/base"
	"github.com/shalb/cluster.dev/pkg/utils"
)

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
	if u.StateData != nil {
		return u.StateData
	}
	return u.GetStateUnit()
}

type UnitDiffSpec struct {
	base.UnitDiffSpec
	Source   string        `json:"source"`
	HelmOpts interface{}   `json:"helm_opts,omitempty"`
	Sets     interface{}   `json:"sets,omitempty"`
	Values   []interface{} `json:"values,omitempty"`
}

func (u *Unit) GetUnitDiff() UnitDiffSpec {
	diff := u.Unit.GetUnitDiff()
	st := UnitDiffSpec{
		UnitDiffSpec: diff,
		Source:       u.Source,
		HelmOpts:     u.HelmOpts,
		Sets:         u.Sets,
	}
	filesListDiff := make([]interface{}, len(u.ValuesFilesList))
	st.Outputs = nil
	for i, str := range u.ValuesFilesList {
		fileLines := strings.Split(str, "\n")
		if len(fileLines) < 2 {
			filesListDiff[i] = str
		} else {
			for _, line := range fileLines {
				if line == "" {
					continue // Ignore empty lines
				}
				filesListDiff = append(filesListDiff, line)
			}
		}
	}
	st.Values = filesListDiff
	return st
}

func (u *Unit) GetDiffData() interface{} {
	st := u.GetUnitDiff()
	res := map[string]interface{}{}
	utils.JSONCopy(st, &res)
	project.ScanMarkers(res, base.StringRemStScanner, u)
	project.ScanMarkers(res, project.OutputsReplacer, u)
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
	return err
}
