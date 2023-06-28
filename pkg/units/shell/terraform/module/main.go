package tfmodule

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"

	"github.com/apex/log"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/shalb/cluster.dev/pkg/config"
	"github.com/shalb/cluster.dev/pkg/hcltools"
	"github.com/shalb/cluster.dev/pkg/project"
	"github.com/shalb/cluster.dev/pkg/units/shell/common"
	"github.com/shalb/cluster.dev/pkg/units/shell/terraform/base"
	"github.com/shalb/cluster.dev/pkg/utils"
	"github.com/zclconf/go-cty/cty"
)

type Unit struct {
	base.Unit
	Source               string                 `yaml:"-" json:"source"`
	Version              string                 `yaml:"-" json:"version,omitempty"`
	Inputs               map[string]interface{} `yaml:"-" json:"inputs,omitempty"`
	LocalModule          *common.FilesListT     `yaml:"-" json:"local_module"`
	CustomFiles          *common.FilesListT     `yaml:"create_files,omitempty" json:"create_files,omitempty"`
	UnitKind             string                 `yaml:"-" json:"type"`
	StateData            interface{}            `yaml:"-" json:"-"`
	LocalModuleCachePath string                 `yaml:"-" json:"-"`
}

func (u *Unit) KindKey() string {
	return unitKind
}

func (u *Unit) genMainCodeBlock() ([]byte, error) {

	f := hclwrite.NewEmptyFile()
	rootBody := f.Body()

	unitBlock := rootBody.AppendNewBlock("module", []string{u.Name()})
	unitBody := unitBlock.Body()
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
	if u.LocalModule != nil {
		relPath, err := filepath.Rel(u.CacheDir, u.LocalModuleCachePath)
		if err != nil {
			return nil, fmt.Errorf("tfModule unit: genMainCodeBlock: %w", err)
		}
		unitBody.SetAttributeValue("source", cty.StringVal(relPath))
	} else {
		unitBody.SetAttributeValue("source", cty.StringVal(u.Source))
	}

	for hash, marker := range u.ProjectPtr.UnitLinks.ByLinkTypes(base.RemoteStateLinkType).Map() {
		if marker.TargetStackName == "this" {
			marker.TargetStackName = u.Stack().Name
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
	uniqMap := map[string]bool{}
	for _, link := range u.ProjectPtr.UnitLinks.ByTargetUnit(u).ByLinkTypes(base.RemoteStateLinkType, project.OutputLinkType).Map() {
		// log.Warnf("     output: %v --> %v", u.Name(), link.OutputName)
		re := regexp.MustCompile(`^[A-Za-z][a-zA-Z0-9_\-]{0,}`)
		outputName := re.FindString(link.OutputName)
		// Deduplicate outputs.
		if uniqMap[outputName] {
			continue
		}
		if len(link.OutputName) < 1 {
			return nil, fmt.Errorf("invalid output '%v' in unit '%v'", link, u.Name())
		}
		dataBlock := rootBody.AppendNewBlock("output", []string{outputName})
		dataBody := dataBlock.Body()
		outputStr := fmt.Sprintf("module.%s.%s", u.Name(), outputName)
		dataBody.SetAttributeRaw("value", hcltools.CreateTokensForOutput(outputStr))
		dataBody.SetAttributeValue("sensitive", cty.BoolVal(true))
		uniqMap[outputName] = true
	}
	return f.Bytes(), nil

}

func (u *Unit) ReadConfig(spec map[string]interface{}, stack *project.Stack) error {
	u.UnitKind = u.KindKey()
	source, ok := spec["source"].(string)
	if !ok {
		return fmt.Errorf("incorrect unit source")
	}
	if utils.IsLocalPath(source) {
		u.LocalModule = &common.FilesListT{}
    var tfModuleLocalDir string
    if utils.IsAbsolutePath(source) {
      tfModuleLocalDir = source
    } else {
      tfModuleLocalDir = filepath.Join(config.Global.WorkingDir, u.Stack().TemplateDir, source)
    }
		err := u.LocalModule.ReadDir(tfModuleLocalDir, tfModuleLocalDir)
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

func (u *Unit) ScanData(scanner project.MarkerScanner) error {
	err := project.ScanMarkers(u.Inputs, scanner, u)
	if err != nil {
		return err
	}
	return nil
}

// Prepare scan all markers in unit, and build project unit links, and unit dependencies.
func (u *Unit) Prepare() error {

	err := u.Unit.Prepare()
	if err != nil {
		return err
	}
	err = u.ScanData(project.OutputsScanner)
	if err != nil {
		return err
	}
	err = u.ScanData(u.RemoteStatesScanner)
	if err != nil {
		return err
	}
	return nil
}

// Build generate all terraform code for project.
func (u *Unit) Build() error {
	u.SavedState = u.GetState()
	err := u.ScanData(project.OutputsReplacer)
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
	if !u.ProjectPtr.UnitLinks.ByTargetUnit(u).IsEmpty() {
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
		err := os.MkdirAll(u.LocalModuleCachePath, 0755)
		if err != nil {
			log.Debugf("save local tf module to cache: mkdir '%v': '%v'", u.LocalModuleCachePath, err.Error())
		}
		err = u.LocalModule.WriteFiles(u.LocalModuleCachePath)
		if err != nil {
			return err
		}
	}
	return nil
}

// UpdateProjectRuntimeData update project runtime, dataset, adds module outputs.
func (u *Unit) UpdateProjectRuntimeData(p *project.Project) error {
	return u.Unit.UpdateProjectRuntimeData(p)
}
