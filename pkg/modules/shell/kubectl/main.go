package tfmodule

import (
	"github.com/shalb/cluster.dev/pkg/modules/shell/common"
	"github.com/shalb/cluster.dev/pkg/project"
)

type Unit struct {
	common.Unit
	outputRaw string
	inputs    map[string]interface{}
}

func (m *Unit) KindKey() string {
	return "printer"
}

func (m *Unit) ReadConfig(spec map[string]interface{}, stack *project.Stack) error {
	return nil
}

// ReplaceMarkers replace all templated markers with values.
func (m *Unit) ReplaceMarkers() error {
	err := m.Unit.ReplaceMarkers()
	if err != nil {
		return err
	}
	return nil
}

// CreateCodeDir generate all terraform code for project.
func (m *Unit) Build() error {
	var err error
	err = m.Unit.Build()
	if err != nil {
		return err
	}
	return m.CreateCodeDir()
}

func (m *Unit) Apply() (err error) {
	err = m.Unit.Apply()
	if err != nil {
		return
	}
	return
}

// UpdateProjectRuntimeData update project runtime dataset, adds printer unit outputs.
func (m *Unit) UpdateProjectRuntimeData(p *project.Project) error {
	p.RuntimeDataset.PrintersOutputs = append(p.RuntimeDataset.PrintersOutputs, project.PrinterOutput{Name: m.Key(), Output: m.outputRaw})
	return m.Unit.UpdateProjectRuntimeData(p)
}
