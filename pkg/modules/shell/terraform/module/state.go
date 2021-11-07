package tfmodule

import (
	"encoding/base64"
	"fmt"

	"github.com/shalb/cluster.dev/pkg/modules/shell/terraform/base"
	"github.com/shalb/cluster.dev/pkg/project"
	"github.com/shalb/cluster.dev/pkg/utils"
)

type UnitDiffSpec struct {
	base.UnitDiffSpec
	Source    string            `json:"source"`
	Version   string            `json:"version,omitempty"`
	Inputs    interface{}       `json:"inputs,omitempty"`
	LocalUnit map[string]string `json:"local_module"`
}

func (m *UnitTfModule) GetState() interface{} {
	m.StatePtr.Unit = m.Unit.GetState().(base.Unit)
	return *m.StatePtr
}

func (m *UnitTfModule) GetUnitDiff() UnitDiffSpec {
	diff := m.Unit.GetUnitDiff()
	st := UnitDiffSpec{
		UnitDiffSpec: diff,
		Source:       m.Source,
		Version:      m.Version,
		Inputs:       m.Inputs,
	}
	if m.LocalUnit != nil && utils.IsLocalPath(m.Source) {
		st.LocalUnit = make(map[string]string)
		for dir, file := range m.LocalUnit {
			st.LocalUnit[dir] = base64.StdEncoding.EncodeToString([]byte(file))
		}
	}
	return st
}

func (m *UnitTfModule) GetDiffData() interface{} {
	st := m.GetUnitDiff()
	res := map[string]interface{}{}
	utils.JSONCopy(st, &res)
	project.ScanMarkers(res, base.StateRemStScanner, m)
	return res
}

func (m *UnitTfModule) LoadState(stateData interface{}, modKey string, p *project.StateProject) error {
	err := m.Unit.LoadState(stateData, modKey, p)
	if err != nil {
		return err
	}
	err = utils.JSONCopy(stateData, m)
	if err != nil {
		return fmt.Errorf("load state: %v", err.Error())
	}
	m.StatePtr = &UnitTfModule{
		Unit: m.Unit,
	}
	err = utils.JSONCopy(m, m.StatePtr)
	for dir, file := range m.LocalUnit {
		m.StatePtr.LocalUnit[dir] = base64.StdEncoding.EncodeToString([]byte(file))
	}
	return err
}
