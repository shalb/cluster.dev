package tfmodule

import (
	"github.com/shalb/cluster.dev/pkg/modules/shell/common"
	"github.com/shalb/cluster.dev/pkg/project"
)

type Module struct {
	common.Unit
	outputRaw string
	inputs    map[string]interface{}
}

func (m *Module) KindKey() string {
	return "printer"
}

func (m *Module) ReadConfig(spec map[string]interface{}, stack *project.Stack) error {
	return nil
}

// ReplaceMarkers replace all templated markers with values.
func (m *Module) ReplaceMarkers() error {
	err := m.Unit.ReplaceMarkers()
	if err != nil {
		return err
	}
	return nil
}

// CreateCodeDir generate all terraform code for project.
func (m *Module) Build() error {
	var err error
	err = m.Unit.Build()
	if err != nil {
		return err
	}
	return m.CreateCodeDir()
}

func (m *Module) Apply() (err error) {
	err = m.Unit.Apply()
	if err != nil {
		return
	}
	return
}

// UpdateProjectRuntimeData update project runtime dataset, adds printer module outputs.
func (m *Module) UpdateProjectRuntimeData(p *project.Project) error {
	p.RuntimeDataset.PrintersOutputs = append(p.RuntimeDataset.PrintersOutputs, project.PrinterOutput{Name: m.Key(), Output: m.outputRaw})
	return m.Unit.UpdateProjectRuntimeData(p)
}
