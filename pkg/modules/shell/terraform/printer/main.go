package tfmodule

import (
	"fmt"
	"io/fs"

	"github.com/apex/log"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/shalb/cluster.dev/pkg/hcltools"
	"github.com/shalb/cluster.dev/pkg/modules/shell/terraform/base"
	"github.com/shalb/cluster.dev/pkg/project"
)

type Unit struct {
	base.Unit
	outputRaw string
	inputs    map[string]interface{}
}

func (m *Unit) KindKey() string {
	return "printer"
}

func (m *Unit) genMainCodeBlock() ([]byte, error) {
	f := hclwrite.NewEmptyFile()
	rootBody := f.Body()

	for key, val := range m.inputs {
		dataBlock := rootBody.AppendNewBlock("output", []string{key})
		dataBody := dataBlock.Body()
		hclVal, err := hcltools.InterfaceToCty(val)
		if err != nil {
			return nil, err
		}
		dataBody.SetAttributeValue("value", hclVal)
		for hash, m := range m.Markers() {
			marker, ok := m.(*project.DependencyOutput)
			// log.Warnf("kubernetes marker printer: %v", marker)
			refStr := base.DependencyToRemoteStateRef(marker)
			if !ok {
				return nil, fmt.Errorf("generate main.tf: internal error: incorrect remote state type")
			}
			hcltools.ReplaceStingMarkerInBody(dataBody, hash, refStr)
		}
	}
	return f.Bytes(), nil
}

func (m *Unit) ReadConfig(spec map[string]interface{}, stack *project.Stack) error {
	err := m.Unit.ReadConfig(spec, stack)
	if err != nil {
		log.Debug(err.Error())
		return err
	}
	modType, ok := spec["type"].(string)
	if !ok {
		return fmt.Errorf("Incorrect unit type")
	}
	if modType != m.KindKey() {
		return fmt.Errorf("Incorrect unit type")
	}
	mInputs, ok := spec["inputs"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("Incorrect unit inputs")
	}
	m.inputs = mInputs
	return nil
}

// ReplaceMarkers replace all templated markers with values.
func (m *Unit) ReplaceMarkers() error {
	err := m.Unit.ReplaceMarkers(m)
	if err != nil {
		return err
	}
	err = project.ScanMarkers(m.inputs, m.RemoteStatesScanner, m)
	if err != nil {
		return err
	}
	return nil
}

// Build generate all terraform code for project.
func (m *Unit) Build() error {
	mainBlock, err := m.genMainCodeBlock()
	if err != nil {
		log.Debug(err.Error())
		return err
	}
	if err = m.CreateFiles.Add("main.tf", string(mainBlock), fs.ModePerm); err != nil {
		return err
	}

	m.CreateFiles.Delete("init.tf")
	return m.Unit.Build()
}

func (m *Unit) Apply() (err error) {
	err = m.Unit.Apply()
	if err != nil {
		return
	}
	outputs, err := m.Output()
	if err != nil {
		return
	}
	m.outputRaw = outputs
	return
}

// UpdateProjectRuntimeData update project runtime dataset, adds printer unit outputs.
func (m *Unit) UpdateProjectRuntimeData(p *project.Project) error {
	p.RuntimeDataset.PrintersOutputs = append(p.RuntimeDataset.PrintersOutputs, project.PrinterOutput{Name: m.Key(), Output: m.outputRaw})
	return m.Unit.UpdateProjectRuntimeData(p)
}
