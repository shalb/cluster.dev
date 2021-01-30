package tfmodule

import (
	"fmt"
	"regexp"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/shalb/cluster.dev/pkg/hcltools"
	"github.com/shalb/cluster.dev/pkg/modules/terraform/common"
	"github.com/shalb/cluster.dev/pkg/project"
	"github.com/zclconf/go-cty/cty"
)

type tfModule struct {
	common.Module
	source string
	inputs map[string]interface{}
}

func (m *tfModule) ModKindKey() string {
	return "terraform"
}

func (m *tfModule) GenMainCodeBlock(mod project.Module) ([]byte, error) {

	f := hclwrite.NewEmptyFile()
	rootBody := f.Body()

	moduleBlock := rootBody.AppendNewBlock("module", []string{m.Name()})
	moduleBody := moduleBlock.Body()
	moduleBody.SetAttributeValue("source", cty.StringVal(m.source))
	for key, val := range m.inputs {
		ctyVal, err := hcltools.InterfaceToCty(val)
		if err != nil {
			return nil, err
		}
		moduleBody.SetAttributeValue(key, ctyVal)
	}
	depMarkers, ok := m.ProjectPtr().Markers["remoteStateMarkerCatName"]
	if ok {
		for hash, marker := range depMarkers.(map[string]*project.Dependency) {
			if marker.Module == nil {
				continue
			}
			remoteStateRef := fmt.Sprintf("data.terraform_remote_state.%s-%s.outputs.%s", marker.Module.InfraName(), marker.Module.Name(), marker.Output)
			hcltools.ReplaceStingMarkerInBody(moduleBody, hash, remoteStateRef)
		}
	}
	return f.Bytes(), nil
}

// genOutputsBlock generate output code block for this module.
func (m *tfModule) GenOutputs(mod project.Module) ([]byte, error) {
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

func (m *tfModule) ReadConfig(spec map[string]interface{}) error {
	modType, ok := spec["type"].(string)
	if !ok {
		return fmt.Errorf("Incorrect module type")
	}
	if modType != m.ModKindKey() {
		return fmt.Errorf("Incorrect module type")
	}
	source, ok := spec["source"].(string)
	if !ok {
		return fmt.Errorf("Incorrect module source")
	}
	m.source = source
	mInputs, ok := spec["inputs"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("Incorrect module inputs")
	}
	m.inputs = mInputs
	m.source = source
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
