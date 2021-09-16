package tfmodule

import (
	"fmt"

	"github.com/apex/log"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/shalb/cluster.dev/pkg/hcltools"
	"github.com/shalb/cluster.dev/pkg/modules/terraform/common"
	"github.com/shalb/cluster.dev/pkg/project"
)

type Module struct {
	common.Module
	outputRaw string
	inputs    map[string]interface{}
}

func (m *Module) KindKey() string {
	return "printer"
}

func (m *Module) genMainCodeBlock() ([]byte, error) {
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
			refStr := common.DependencyToRemoteStateRef(marker)
			if !ok {
				return nil, fmt.Errorf("generate main.tf: internal error: incorrect remote state type")
			}
			hcltools.ReplaceStingMarkerInBody(dataBody, hash, refStr)
		}
	}
	return f.Bytes(), nil
}

func (m *Module) ReadConfig(spec map[string]interface{}, infra *project.Infrastructure) error {
	err := m.ReadConfigCommon(spec, infra)
	if err != nil {
		log.Debug(err.Error())
		return err
	}
	modType, ok := spec["type"].(string)
	if !ok {
		return fmt.Errorf("Incorrect module type")
	}
	if modType != m.KindKey() {
		return fmt.Errorf("Incorrect module type")
	}
	mInputs, ok := spec["inputs"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("Incorrect module inputs")
	}
	m.inputs = mInputs
	return nil
}

// ReplaceMarkers replace all templated markers with values.
func (m *Module) ReplaceMarkers() error {
	err := m.ReplaceMarkersCommon(m)
	if err != nil {
		return err
	}
	err = project.ScanMarkers(m.inputs, project.YamlBlockMarkerScanner, m)
	if err != nil {
		return err
	}
	err = project.ScanMarkers(m.inputs, m.RemoteStatesScanner, m)
	if err != nil {
		return err
	}
	return nil
}

// CreateCodeDir generate all terraform code for project.
func (m *Module) Build() error {
	var err error
	err = m.BuildCommon()
	if err != nil {
		return err
	}
	m.FilesList()["main.tf"], err = m.genMainCodeBlock()
	if err != nil {
		log.Debug(err.Error())
		return err
	}

	// if len(m.ExpectedOutputs()) > 0 {
	// 	return fmt.Errorf("module type 'printer' cannot have outputs, don't use remote state to it")
	// }
	// Remove backend for printer.
	delete(m.FilesList(), "init.tf")
	return m.CreateCodeDir()
}

func (m *Module) Apply() (err error) {
	err = m.ApplyCommon()
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

// UpdateProjectRuntimeData update project runtime dataset, adds printer module outputs.
func (m *Module) UpdateProjectRuntimeData(p *project.Project) error {
	p.RuntimeDataset.PrintersOutputs = append(p.RuntimeDataset.PrintersOutputs, project.PrinterOutput{Name: m.Key(), Output: m.outputRaw})
	return m.UpdateProjectRuntimeDataCommon(p)
}
