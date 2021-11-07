package tfmodule

import (
	"encoding/base64"
	"fmt"
	"io/fs"
	"path/filepath"
	"regexp"

	"github.com/apex/log"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/shalb/cluster.dev/pkg/config"
	"github.com/shalb/cluster.dev/pkg/hcltools"
	"github.com/shalb/cluster.dev/pkg/modules/shell/terraform/base"
	"github.com/shalb/cluster.dev/pkg/project"
	"github.com/shalb/cluster.dev/pkg/utils"
	"github.com/zclconf/go-cty/cty"
)

type UnitTfModule struct {
	base.Unit
	Source    string                 `yaml:"-" json:"source"`
	Version   string                 `yaml:"-" json:"version,omitempty"`
	Inputs    map[string]interface{} `yaml:"-" json:"inputs,omitempty"`
	LocalUnit map[string]string      `yaml:"-" json:"local_module"`
	StatePtr  *UnitTfModule          `yaml:"-" json:"-"`
	UnitKind  string                 `yaml:"-" json:"type"`
}

func (m *UnitTfModule) KindKey() string {
	return unitKind
}

func (m *UnitTfModule) genMainCodeBlock() ([]byte, error) {

	f := hclwrite.NewEmptyFile()
	rootBody := f.Body()

	unitBlock := rootBody.AppendNewBlock("module", []string{m.Name()})
	unitBody := unitBlock.Body()
	unitBody.SetAttributeValue("source", cty.StringVal(m.Source))
	if m.Version != "" {
		unitBody.SetAttributeValue("version", cty.StringVal(m.Version))
	}

	for key, val := range m.Inputs {
		ctyVal, err := hcltools.InterfaceToCty(val)
		if err != nil {
			return nil, err
		}
		unitBody.SetAttributeValue(key, ctyVal)
	}
	if utils.IsLocalPath(m.Source) {
		log.Debugf("Writing local tf unit files to %v unit dir", m.Key())
		err := utils.WriteFilesFromList(m.CodeDir(), m.LocalUnit)
		if err != nil {
			return nil, fmt.Errorf("%v, reading local unit: %v", m.Key(), err.Error())
		}
		unitBody.SetAttributeValue("source", cty.StringVal(m.Source))
	}
	for hash, m := range m.Markers() {
		marker, ok := m.(*project.DependencyOutput)
		if !ok {
			return nil, fmt.Errorf("generate main.tf: internal error: incorrect remote state type")
		}
		refStr := base.DependencyToRemoteStateRef(marker)
		hcltools.ReplaceStingMarkerInBody(unitBody, hash, refStr)
	}
	return f.Bytes(), nil
}

// genOutputsBlock generate output code block for this unit.
func (m *UnitTfModule) genOutputs() ([]byte, error) {
	f := hclwrite.NewEmptyFile()

	rootBody := f.Body()
	for output := range m.ExpectedOutputs() {
		re := regexp.MustCompile(`^[A-Za-z][a-zA-Z0-9_\-]{0,}`)
		outputName := re.FindString(output)
		if len(outputName) < 1 {
			return nil, fmt.Errorf("invalid output '%v' in unit '%v'", output, m.Name())
		}
		dataBlock := rootBody.AppendNewBlock("output", []string{outputName})
		dataBody := dataBlock.Body()
		outputStr := fmt.Sprintf("module.%s.%s", m.Name(), outputName)
		dataBody.SetAttributeRaw("value", hcltools.CreateTokensForOutput(outputStr))
	}
	return f.Bytes(), nil

}

func (m *UnitTfModule) ReadConfig(spec map[string]interface{}, stack *project.Stack) error {
	modType, ok := spec["type"].(string)
	if !ok {
		return fmt.Errorf("Incorrect unit type")
	}
	if modType != m.KindKey() {
		return fmt.Errorf("Incorrect unit type")
	}
	m.UnitKind = m.KindKey()
	source, ok := spec["source"].(string)
	if !ok {
		return fmt.Errorf("Incorrect unit source")
	}
	if utils.IsLocalPath(source) {
		tfModuleLocalDir := filepath.Join(config.Global.WorkingDir, m.Stack().TemplateDir, source)
		tfModuleBasePath := filepath.Join(config.Global.WorkingDir, m.Stack().TemplateDir)

		log.Debugf("Reading local tf unit files %v %v ", tfModuleLocalDir, tfModuleBasePath)

		err := m.CreateFiles.ReadDir(tfModuleLocalDir, tfModuleBasePath)
		if err != nil {
			return fmt.Errorf("%v, reading local unit: %v", m.Key(), err.Error())
		}
	}
	if version, ok := spec["version"]; ok {
		m.Version = fmt.Sprintf("%v", version)
	}
	m.Source = source
	mInputs, ok := spec["inputs"].(map[string]interface{})
	if !ok {
		mInputs = nil
	}
	m.Inputs = mInputs
	m.StatePtr = &UnitTfModule{
		Unit: m.Unit,
	}
	err := utils.JSONCopy(m, m.StatePtr)
	for dir, file := range m.LocalUnit {
		m.StatePtr.LocalUnit[dir] = base64.StdEncoding.EncodeToString([]byte(file))
	}
	return err
}

// ReplaceMarkers replace all templated markers with values.
func (m *UnitTfModule) ReplaceMarkers() error {
	err := m.Unit.ReplaceMarkers()
	if err != nil {
		return err
	}
	// log.Warnf("%+v", m.inputs)
	err = project.ScanMarkers(m.Inputs, m.RemoteStatesScanner, m)
	if err != nil {
		return err
	}
	return nil
}

// Build generate all terraform code for project.
func (m *UnitTfModule) Build() error {

	mainFile, err := m.genMainCodeBlock()
	if err != nil {
		log.Debug(err.Error())
		return err
	}
	err = m.CreateFiles.Add("main.tf", string(mainFile), fs.ModePerm)
	if err != nil {
		log.Debug(err.Error())
		return err
	}
	if len(m.ExpectedOutputs()) > 0 {
		outputsFile, err := m.genOutputs()
		if err != nil {
			log.Debug(err.Error())
			return err
		}
		err = m.CreateFiles.Add("outputs.tf", string(outputsFile), fs.ModePerm)
		if err != nil {
			log.Debug(err.Error())
			return err
		}
	}

	return m.Unit.Build()
}

// UpdateProjectRuntimeData update project runtime dataset, adds module outputs.
func (m *UnitTfModule) UpdateProjectRuntimeData(p *project.Project) error {
	return m.Unit.UpdateProjectRuntimeData(p)
}
