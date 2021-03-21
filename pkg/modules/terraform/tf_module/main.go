package tfmodule

import (
	"fmt"
	"regexp"

	"github.com/apex/log"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/shalb/cluster.dev/pkg/hcltools"
	"github.com/shalb/cluster.dev/pkg/modules/terraform/common"
	"github.com/shalb/cluster.dev/pkg/project"
	"github.com/zclconf/go-cty/cty"
)

type tfModule struct {
	common.Module
	source  string
	version string
	inputs  map[string]interface{}
}

func (m *tfModule) KindKey() string {
	return "terraform"
}

func (m *tfModule) genMainCodeBlock() ([]byte, error) {

	f := hclwrite.NewEmptyFile()
	rootBody := f.Body()

	moduleBlock := rootBody.AppendNewBlock("module", []string{m.Name()})
	moduleBody := moduleBlock.Body()
	moduleBody.SetAttributeValue("source", cty.StringVal(m.source))
	if m.version != "" {
		moduleBody.SetAttributeValue("version", cty.StringVal(m.version))
	}
	for key, val := range m.inputs {
		ctyVal, err := hcltools.InterfaceToCty(val)
		if err != nil {
			return nil, err
		}
		moduleBody.SetAttributeValue(key, ctyVal)
	}
	for hash, ref := range m.Markers() {
		hcltools.ReplaceStingMarkerInBody(moduleBody, hash, ref)
	}
	return f.Bytes(), nil
}

// genOutputsBlock generate output code block for this module.
func (m *tfModule) genOutputs() ([]byte, error) {
	f := hclwrite.NewEmptyFile()

	rootBody := f.Body()
	for output := range m.ExpectedOutputs() {
		re := regexp.MustCompile(`^[A-Za-z][a-zA-Z0-9_\-]{0,}`)
		outputName := re.FindString(output)
		if len(outputName) < 1 {
			return nil, fmt.Errorf("invalid output '%v' in module '%v'", output, m.Name())
		}
		dataBlock := rootBody.AppendNewBlock("output", []string{outputName})
		dataBody := dataBlock.Body()
		outputStr := fmt.Sprintf("module.%s.%s", m.Name(), outputName)
		dataBody.SetAttributeRaw("value", hcltools.CreateTokensForOutput(outputStr))
	}
	return f.Bytes(), nil

}

func (m *tfModule) ReadConfig(spec map[string]interface{}, infra *project.Infrastructure) error {
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
	source, ok := spec["source"].(string)
	if !ok {
		return fmt.Errorf("Incorrect module source")
	}
	if version, ok := spec["version"]; ok {
		m.version = fmt.Sprintf("%v", version)
	}
	m.source = source
	mInputs, ok := spec["inputs"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("Incorrect module inputs")
	}
	m.inputs = mInputs
	return nil
}

// ReplaceMarkers replace all templated markers with values.
func (m *tfModule) ReplaceMarkers() error {
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
func (m *tfModule) Build(codeDir string) error {
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
		m.FilesList()["outputs.tf"], err = m.genOutputs()
		if err != nil {
			log.Debug(err.Error())
			return err
		}
	}
	return m.CreateCodeDir(codeDir)
}
