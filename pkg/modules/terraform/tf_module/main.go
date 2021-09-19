package tfmodule

import (
	"fmt"
	"path/filepath"
	"regexp"

	"github.com/apex/log"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/shalb/cluster.dev/pkg/config"
	"github.com/shalb/cluster.dev/pkg/hcltools"
	"github.com/shalb/cluster.dev/pkg/modules/terraform/common"
	"github.com/shalb/cluster.dev/pkg/project"
	"github.com/shalb/cluster.dev/pkg/utils"
	"github.com/zclconf/go-cty/cty"
)

type Module struct {
	common.Unit
	source      string
	version     string
	inputs      map[string]interface{}
	localModule map[string][]byte
}

func (m *Module) KindKey() string {
	return "terraform"
}

func (m *Module) genMainCodeBlock() ([]byte, error) {

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
	if utils.IsLocalPath(m.source) {
		log.Debugf("Writing local tf module files to %v module dir", m.Key())
		err := utils.WriteFilesFromList(m.CodeDir(), m.localModule)
		if err != nil {
			return nil, fmt.Errorf("%v, reading local module: %v", m.Key(), err.Error())
		}
		moduleBody.SetAttributeValue("source", cty.StringVal(m.source))
	}
	for hash, m := range m.Markers() {
		marker, ok := m.(*project.DependencyOutput)
		if !ok {
			return nil, fmt.Errorf("generate main.tf: internal error: incorrect remote state type")
		}
		refStr := common.DependencyToRemoteStateRef(marker)
		hcltools.ReplaceStingMarkerInBody(moduleBody, hash, refStr)
	}
	return f.Bytes(), nil
}

// genOutputsBlock generate output code block for this module.
func (m *Module) genOutputs() ([]byte, error) {
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

func (m *Module) ReadConfig(spec map[string]interface{}, stack *project.Stack) error {
	err := m.Unit.ReadConfig(spec, stack)
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
	if utils.IsLocalPath(source) {
		tfModuleLocalDir := filepath.Join(config.Global.WorkingDir, m.StackPtr().TemplateDir, source)
		tfModuleBasePath := filepath.Join(config.Global.WorkingDir, m.StackPtr().TemplateDir)
		var err error
		log.Debugf("Reading local tf module files %v %v ", tfModuleLocalDir, tfModuleBasePath)
		m.localModule, err = utils.ReadFilesToList(tfModuleLocalDir, tfModuleBasePath)
		if err != nil {
			return fmt.Errorf("%v, reading local module: %v", m.Key(), err.Error())
		}
	}
	if version, ok := spec["version"]; ok {
		m.version = fmt.Sprintf("%v", version)
	}
	m.source = source
	mInputs, ok := spec["inputs"].(map[string]interface{})
	if !ok {
		mInputs = nil
	}
	m.inputs = mInputs
	return nil
}

// ReplaceMarkers replace all templated markers with values.
func (m *Module) ReplaceMarkers() error {
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
func (m *Module) Build() error {
	err := m.Unit.Build()
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
	return m.CreateCodeDir()
}

// UpdateProjectRuntimeData update project runtime dataset, adds module outputs.
func (m *Module) UpdateProjectRuntimeData(p *project.Project) error {
	return m.Unit.UpdateProjectRuntimeData(p)
}
