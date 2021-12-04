package tfmodule

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"regexp"

	"github.com/apex/log"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/shalb/cluster.dev/pkg/config"
	"github.com/shalb/cluster.dev/pkg/hcltools"
	"github.com/shalb/cluster.dev/pkg/modules/shell/common"
	"github.com/shalb/cluster.dev/pkg/modules/shell/terraform/base"
	"github.com/shalb/cluster.dev/pkg/project"
	"github.com/shalb/cluster.dev/pkg/utils"
	"github.com/zclconf/go-cty/cty"
)

type Unit struct {
	base.Unit
	Source      string                 `yaml:"-" json:"source"`
	Version     string                 `yaml:"-" json:"version,omitempty"`
	Inputs      map[string]interface{} `yaml:"-" json:"inputs,omitempty"`
	LocalModule *common.FilesListT     `yaml:"-" json:"local_module"`
	UnitKind    string                 `yaml:"-" json:"type"`
}

func (u *Unit) KindKey() string {
	return unitKind
}

func (u *Unit) genMainCodeBlock() ([]byte, error) {

	f := hclwrite.NewEmptyFile()
	rootBody := f.Body()

	unitBlock := rootBody.AppendNewBlock("module", []string{u.Name()})
	unitBody := unitBlock.Body()
	unitBody.SetAttributeValue("source", cty.StringVal(u.Source))
	if u.Version != "" {
		unitBody.SetAttributeValue("version", cty.StringVal(u.Version))
	}

	for key, val := range u.Inputs {
		ctyVal, err := hcltools.InterfaceToCty(val)
		if err != nil {
			return nil, err
		}
		unitBody.SetAttributeValue(key, ctyVal)
	}
	if utils.IsLocalPath(u.Source) {
		unitBody.SetAttributeValue("source", cty.StringVal(u.Source))
	}
	markersList := map[string]*project.DependencyOutput{}
	err := u.Project().GetMarkers(base.RemoteStateMarkerCatName, markersList)
	if err != nil {
		return nil, err
	}
	for hash, marker := range markersList {
		if marker.StackName == "this" {
			marker.StackName = u.Stack().Name
		}
		refStr := base.DependencyToRemoteStateRef(marker)
		hcltools.ReplaceStingMarkerInBody(unitBody, hash, refStr)
	}
	return f.Bytes(), nil
}

// genOutputsBlock generate output code block for this unit.
func (u *Unit) genOutputs() ([]byte, error) {
	f := hclwrite.NewEmptyFile()

	rootBody := f.Body()
	for output := range u.ExpectedOutputs().List {
		re := regexp.MustCompile(`^[A-Za-z][a-zA-Z0-9_\-]{0,}`)
		outputName := re.FindString(output)
		if len(outputName) < 1 {
			return nil, fmt.Errorf("invalid output '%v' in unit '%v'", output, u.Name())
		}
		dataBlock := rootBody.AppendNewBlock("output", []string{outputName})
		dataBody := dataBlock.Body()
		outputStr := fmt.Sprintf("module.%s.%s", u.Name(), outputName)
		dataBody.SetAttributeRaw("value", hcltools.CreateTokensForOutput(outputStr))
	}
	return f.Bytes(), nil

}

func (u *Unit) ReadConfig(spec map[string]interface{}, stack *project.Stack) error {
	u.UnitKind = u.KindKey()
	source, ok := spec["source"].(string)
	if !ok {
		return fmt.Errorf("Incorrect unit source")
	}
	if utils.IsLocalPath(source) {
		u.LocalModule = &common.FilesListT{}
		tfModuleLocalDir := filepath.Join(config.Global.WorkingDir, u.Stack().TemplateDir, source)
		tfModuleBasePath := filepath.Join(config.Global.WorkingDir, u.Stack().TemplateDir)

		log.Debugf("Reading local tf unit files %v %v ", tfModuleLocalDir, tfModuleBasePath)

		err := u.LocalModule.ReadDir(tfModuleLocalDir, tfModuleBasePath)
		if err != nil {
			return fmt.Errorf("%v, reading local unit: %v", u.Key(), err.Error())
		}
	}
	if version, ok := spec["version"]; ok {
		u.Version = fmt.Sprintf("%v", version)
	}
	u.Source = source
	mInputs, ok := spec["inputs"].(map[string]interface{})
	if !ok {
		mInputs = nil
	}
	u.Inputs = mInputs
	return nil
}

// ReplaceMarkers replace all templated markers with values.
func (u *Unit) ReplaceMarkers() error {
	err := u.Unit.ReplaceMarkers()
	if err != nil {
		return err
	}
	// log.Warnf("%+v", m.inputs)
	err = project.ScanMarkers(u.Inputs, u.RemoteStatesScanner, u)
	if err != nil {
		return err
	}
	err = project.ScanMarkers(u.Inputs, project.OutputsScanner, u)
	if err != nil {
		return err
	}
	return nil
}

// Build generate all terraform code for project.
func (u *Unit) Build() error {
	err := u.ReplaceMarkers()
	if err != nil {
		return err
	}
	mainFile, err := u.genMainCodeBlock()
	if err != nil {
		log.Debug(err.Error())
		return err
	}
	err = u.CreateFiles.Add("main.tf", string(mainFile), fs.ModePerm)
	if err != nil {
		log.Debug(err.Error())
		return err
	}
	if len(u.ExpectedOutputs().List) > 0 {
		outputsFile, err := u.genOutputs()
		if err != nil {
			log.Debug(err.Error())
			return err
		}
		err = u.CreateFiles.Add("outputs.tf", string(outputsFile), fs.ModePerm)
		if err != nil {
			log.Debug(err.Error())
			return err
		}
	}
	if err = u.Unit.Build(); err != nil {
		return err
	}
	if u.LocalModule != nil {
		err = u.LocalModule.WriteFiles(u.CacheDir)
		if err != nil {
			return err
		}
	}
	return nil
}

// UpdateProjectRuntimeData update project runtime dataset, adds module outputs.
func (m *Unit) UpdateProjectRuntimeData(p *project.Project) error {
	return m.Unit.UpdateProjectRuntimeData(p)
}
