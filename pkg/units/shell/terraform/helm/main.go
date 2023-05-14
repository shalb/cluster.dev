package helm

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"path/filepath"
	"reflect"

	"github.com/apex/log"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/shalb/cluster.dev/pkg/config"
	"github.com/shalb/cluster.dev/pkg/hcltools"
	"github.com/shalb/cluster.dev/pkg/project"
	"github.com/shalb/cluster.dev/pkg/units/shell/terraform/base"
	"github.com/shalb/cluster.dev/pkg/utils"
	"github.com/zclconf/go-cty/cty"
	"gopkg.in/yaml.v3"
)

type Unit struct {
	base.Unit
	Source          string                   `yaml:"-" json:"source"`
	HelmOpts        map[string]interface{}   `yaml:"-" json:"helm_opts,omitempty"`
	Sets            map[string]interface{}   `yaml:"-" json:"sets,omitempty"`
	Kubeconfig      *string                  `yaml:"-" json:"kubeconfig"`
	ValuesFilesList []string                 `yaml:"-" json:"values,omitempty"`
	ValuesYAML      []map[string]interface{} `yaml:"-" json:"-"`
	UnitKind        string                   `yaml:"-" json:"type"`
	StateData       interface{}              `yaml:"-" json:"-"`
}

func (u *Unit) KindKey() string {
	return unitKind
}

func (u *Unit) genMainCodeBlock() ([]byte, error) {
	// var marker string
	// if len(m.valuesFileContent) > 0 {
	// 	m.FilesList()["values.yaml"] = m.valuesFileContent
	// }
	f := hclwrite.NewEmptyFile()
	providerBody := &hclwrite.Body{}
	rootBody := f.Body()
	if u.Kubeconfig != nil {
		providerBlock := rootBody.AppendNewBlock("provider", []string{"helm"})
		providerBody = providerBlock.Body()
		provederKubernetesBlock := providerBody.AppendNewBlock("kubernetes", []string{})
		provederKubernetesBlock.Body().SetAttributeValue("config_path", cty.StringVal(*u.Kubeconfig))
		if config.Global.LogLevel == "debug" {
			providerBody.SetAttributeValue("debug", cty.BoolVal(true))
		}
	}
	helmBlock := rootBody.AppendNewBlock("resource", []string{"helm_release", project.ConvertToTfVarName(u.Name())})
	helmBody := helmBlock.Body()
	for key, val := range u.HelmOpts {
		ctyVal, err := hcltools.InterfaceToCty(val)
		if err != nil {
			return nil, err
		}
		helmBody.SetAttributeValue(key, ctyVal)
	}
	for key, val := range u.Sets {
		ctyVal, err := hcltools.InterfaceToCty(val)
		if err != nil {
			return nil, err
		}
		setBlock := helmBody.AppendNewBlock("set", []string{})
		setBlock.Body().SetAttributeValue("name", cty.StringVal(key))
		setBlock.Body().SetAttributeValue("value", ctyVal)
	}
	if len(u.ValuesFilesList) > 0 {
		ctyValuesList := []cty.Value{}
		for _, v := range u.ValuesFilesList {
			ctyValuesList = append(ctyValuesList, cty.StringVal(string(v)))
		}
		helmBody.SetAttributeValue("values", cty.ListVal(ctyValuesList))
		//hcltools.ReplaceStingMarkerInBody(helmBody, marker, "file(\"./values.yaml\")")
	}
	for hash, marker := range u.ProjectPtr.UnitLinks.ByLinkTypes(base.RemoteStateLinkType).Map() {
		if marker.TargetStackName == "this" {
			marker.TargetStackName = u.Stack().Name
		}
		refStr := base.DependencyToRemoteStateRef(marker)
		hcltools.ReplaceStingMarkerInBody(helmBody, hash, refStr)
		hcltools.ReplaceStingMarkerInBody(providerBody, hash, refStr)
	}
	return f.Bytes(), nil
}

func (u *Unit) ReadConfig(spec map[string]interface{}, stack *project.Stack) error {

	source, ok := spec["source"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("read unit config: incorrect unit source, %v", u.Key())
	}
	for key, val := range source {
		u.HelmOpts[key] = val
	}
	helmChartOpt, exists := u.HelmOpts["chart"].(string)
	if !exists {
		return fmt.Errorf("read helm chart configuration: option 'chart' is required and should be a string")
	}
	if utils.IsLocalPath(helmChartOpt) {
		if !utils.IsAbsolutePath(helmChartOpt) {
			absoluteChartPath := filepath.Join(config.Global.ProjectConfigsPath, u.StackPtr.TemplateDir, helmChartOpt)
			u.HelmOpts["chart"] = absoluteChartPath
		}
	}
	kubeconfig, ok := spec["kubeconfig"].(string)
	// if !ok {
	// 	return fmt.Errorf("incorrect kubeconfig")
	// }
	if ok {
		u.Kubeconfig = &kubeconfig
	}
	addOp, ok := spec["additional_options"].(map[string]interface{})
	if ok {
		for key, val := range addOp {
			u.HelmOpts[key] = val
		}
	}
	sets, ok := spec["inputs"].(map[string]interface{})
	if ok {
		for key, val := range sets {
			u.Sets[key] = val
		}
	}
	u.HelmOpts["name"] = spec["name"]
	valuesCat, ok := spec["values"]
	if ok {
		valuesCatList, check := valuesCat.([]interface{})
		if !check {
			return fmt.Errorf("read unit config: 'values' have unknown type: %v", reflect.TypeOf(valuesCat))
		}
		u.ValuesFilesList = []string{}
		//log.Warnf("%v", ok)

		for _, valuesCat := range valuesCatList {
			valuesCatMap, check := valuesCat.(map[string]interface{})
			if !check {
				return fmt.Errorf("read unit config: 'values' have unknown format: %v", reflect.TypeOf(valuesCat))
			}
			applyTemplate, exists := valuesCatMap["apply_template"].(bool)
			if !exists {
				applyTemplate = true
			}
			valuesFileName, ok := valuesCatMap["file"].(string)
			if !ok {
				return fmt.Errorf("read unit config: 'values.file' is required field")
			}
			vfPath := filepath.Join(u.Stack().TemplateDir, valuesFileName)
			valuesFileContent, err := ioutil.ReadFile(vfPath)
			if err != nil {
				log.Debugf(err.Error())
				return fmt.Errorf("read unit config: can't load values file: %v", err)
			}
			values := valuesFileContent
			if applyTemplate {
				renderedValues, errIsWarn, err := u.Stack().TemplateTry(valuesFileContent, vfPath)
				if err != nil {
					if !errIsWarn {
						log.Fatal(err.Error())
					}
				}
				values = renderedValues
			}
			vYAML := make(map[string]interface{})
			err = yaml.Unmarshal(values, &vYAML)
			if err != nil {
				return fmt.Errorf("read unit config: unmarshal values file: %v", utils.ResolveYamlError(values, err))
			}
			u.ValuesYAML = append(u.ValuesYAML, vYAML)
			u.ValuesFilesList = append(u.ValuesFilesList, string(values))
		}
	}
	pv, ok := spec["provider_version"].(string)
	if ok {
		u.AddRequiredProvider("helm", "hashicorp/helm", pv)
	}
	return nil
}

func (u *Unit) ScanData(scanner project.MarkerScanner) error {
	err := project.ScanMarkers(u.ValuesFilesList, scanner, u)
	if err != nil {
		return err
	}
	err = project.ScanMarkers(u.HelmOpts, scanner, u)
	if err != nil {
		return err
	}
	err = project.ScanMarkers(u.Sets, scanner, u)
	if err != nil {
		return err
	}
	err = project.ScanMarkers(u.Providers, scanner, u)
	log.Warnf("ScanData %v", u.Providers)
	if err != nil {
		return err
	}
	err = project.ScanMarkers(u.ValuesYAML, scanner, u)
	if err != nil {
		return err
	}
	err = project.ScanMarkers(u.Kubeconfig, scanner, u)
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
	mainBlock, err := u.genMainCodeBlock()
	if err != nil {
		log.Debug(err.Error())
		return err
	}
	if err = u.CreateFiles.Add("main.tf", string(mainBlock), fs.ModePerm); err != nil {
		return err
	}
	return u.Unit.Build()

}

// UpdateProjectRuntimeData update project runtime dataset, adds unit outputs.
func (u *Unit) UpdateProjectRuntimeData(p *project.Project) error {
	return u.Unit.UpdateProjectRuntimeData(p)
}
