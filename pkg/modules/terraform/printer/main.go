package tfmodule

import (
	"fmt"

	"github.com/apex/log"
	"github.com/gookit/color"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/shalb/cluster.dev/pkg/hcltools"
	"github.com/shalb/cluster.dev/pkg/modules/terraform/common"
	"github.com/shalb/cluster.dev/pkg/project"
)

type printer struct {
	common.Module
	inputs map[string]interface{}
}

func (m *printer) KindKey() string {
	return "printer"
}

func (m *printer) genMainCodeBlock() ([]byte, error) {
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
		for hash, ref := range m.Markers() {
			hcltools.ReplaceStingMarkerInBody(dataBody, hash, ref)
		}
	}
	return f.Bytes(), nil
}

func (m *printer) ReadConfig(spec map[string]interface{}, infra *project.Infrastructure) error {
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
func (m *printer) ReplaceMarkers() error {
	err := project.ScanMarkers(m.inputs, m.YamlBlockMarkerScanner, m)
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
func (m *printer) Build(codeDir string) error {
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

	if len(m.ExpectedOutputs()) > 0 {
		return fmt.Errorf("module type 'printer' cannot have outputs, don't use remote state to it")
	}
	// Remove backend for printer.
	delete(m.FilesList(), "init.tf")
	return m.CreateCodeDir(codeDir)
}

func (m *printer) Apply() (err error) {
	err = m.ApplyCommon()
	if err != nil {
		return
	}
	outputs, err := m.Outputs()
	if err != nil {
		return
	}

	log.Infof("Printer output. Module: '%v', Infra: '%v'\n%v", m.Name(), m.InfraName(), color.Style{color.FgGreen, color.OpBold}.Sprintf(outputs))

	return
}
