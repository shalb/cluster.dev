package helm

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"

	"github.com/apex/log"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/shalb/cluster.dev/internal/config"
	"github.com/shalb/cluster.dev/internal/project"
	"github.com/shalb/cluster.dev/internal/units/shell/common"
	"github.com/shalb/cluster.dev/internal/units/shell/terraform/base"
	"github.com/shalb/cluster.dev/internal/units/shell/terraform/types"
	"github.com/shalb/cluster.dev/pkg/hcltools"
	"github.com/shalb/cluster.dev/pkg/utils"
	"github.com/zclconf/go-cty/cty"
	"gopkg.in/yaml.v3"
)

type Unit struct {
	base.Unit
	Source          string                    `yaml:"-,omitempty" json:"source"`
	HelmOpts        map[string]interface{}    `yaml:"-" json:"helm_opts,omitempty"`
	Sets            map[string]interface{}    `yaml:"-" json:"sets,omitempty"`
	Kubeconfig      *string                   `yaml:"-" json:"kubeconfig"`
	ValuesFilesList []string                  `yaml:"-" json:"values,omitempty"`
	ValuesYAML      []map[string]interface{}  `yaml:"-" json:"-"`
	UnitKind        string                    `yaml:"-" json:"type"`
	StateData       project.Unit              `yaml:"-" json:"-"`
	CustomFiles     *common.FilesListT        `yaml:"create_files,omitempty" json:"create_files,omitempty"`
	ProviderConf    *types.ProviderConfigSpec `yaml:"provider_conf" json:"provider_conf"`
	ProviderVersion string                    `yaml:"-" json:"provider_version,omitempty"`
}

func (u *Unit) KindKey() string {
	return unitKind
}

// isHelmProviderV3OrLater determines if the Helm provider version is v3.0.0 or later
// where kubernetes, registry, and experiments are nested objects instead of blocks
func (u *Unit) isHelmProviderV3OrLater() bool {
	if u.ProviderVersion == "" {
		// Default to v3 behavior for new installations
		return true
	}
	
	// Parse version string and check if >= 3.0.0
	// For simplicity, we'll check if version starts with "3." or higher
	if len(u.ProviderVersion) > 0 {
		firstChar := u.ProviderVersion[0]
		if firstChar >= '3' {
			return true
		}
	}
	return false
}

func (u *Unit) genMainCodeBlock() ([]byte, error) {
	// var marker string
	// if len(m.valuesFileContent) > 0 {
	// 	m.FilesList()["values.yaml"] = m.valuesFileContent
	// }
	f := hclwrite.NewEmptyFile()
	providerBody := &hclwrite.Body{}
	rootBody := f.Body()
	if u.ProviderConf != nil || u.Kubeconfig != nil {
		providerBlock := rootBody.AppendNewBlock("provider", []string{"helm"})
		providerBody = providerBlock.Body()
		
		if u.isHelmProviderV3OrLater() {
			// Helm provider v3+: kubernetes is a nested object
			err := u.generateV3ProviderConfig(providerBody)
			if err != nil {
				return nil, fmt.Errorf("generate v3 provider config: %w", err)
			}
		} else {
			// Helm provider v2: kubernetes is a block
			err := u.generateV2ProviderConfig(providerBody)
			if err != nil {
				return nil, fmt.Errorf("generate v2 provider config: %w", err)
			}
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
	// Handle set values with version-specific syntax
	if len(u.Sets) > 0 {
		if u.isHelmProviderV3OrLater() {
			// v3 syntax: set as list of objects
			err := u.generateV3SetValues(helmBody)
			if err != nil {
				return nil, fmt.Errorf("generate v3 set values: %w", err)
			}
		} else {
			// v2 syntax: set as blocks
			err := u.generateV2SetValues(helmBody)
			if err != nil {
				return nil, fmt.Errorf("generate v2 set values: %w", err)
			}
		}
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

// generateV2ProviderConfig generates Helm provider configuration for v2.x (kubernetes as block)
func (u *Unit) generateV2ProviderConfig(providerBody *hclwrite.Body) error {
	providerKubernetesBlock := providerBody.AppendNewBlock("kubernetes", []string{})
	
	if u.ProviderConf != nil {
		providerCty, err := hcltools.InterfaceToCty(*u.ProviderConf)
		if err != nil {
			return err
		}
		for key, val := range providerCty.AsValueMap() {
			providerKubernetesBlock.Body().SetAttributeValue(key, val)
		}
	}
	
	if u.Kubeconfig != nil {
		log.Warn("Deprecation warning: helm unit option 'kubeconfig' is deprecated. Please use 'provider_conf.config_path' instead.")
		providerKubernetesBlock.Body().SetAttributeValue("config_path", cty.StringVal(*u.Kubeconfig))
	}
	
	return nil
}

// generateV3ProviderConfig generates Helm provider configuration for v3.x+ (kubernetes as nested object)
func (u *Unit) generateV3ProviderConfig(providerBody *hclwrite.Body) error {
	// Create kubernetes configuration as a nested object
	kubernetesConfig := make(map[string]cty.Value)
	
	if u.ProviderConf != nil {
		providerCty, err := hcltools.InterfaceToCty(*u.ProviderConf)
		if err != nil {
			return err
		}
		for key, val := range providerCty.AsValueMap() {
			kubernetesConfig[key] = val
		}
	}
	
	if u.Kubeconfig != nil {
		log.Warn("Deprecation warning: helm unit option 'kubeconfig' is deprecated. Please use 'provider_conf.config_path' instead.")
		kubernetesConfig["config_path"] = cty.StringVal(*u.Kubeconfig)
	}
	
	// Convert to cty object and set as nested object
	if len(kubernetesConfig) > 0 {
		kubernetesCty := cty.ObjectVal(kubernetesConfig)
		providerBody.SetAttributeValue("kubernetes", kubernetesCty)
	}
	
	return nil
}

// generateV2SetValues generates set values for Helm provider v2.x (blocks)
func (u *Unit) generateV2SetValues(helmBody *hclwrite.Body) error {
	for key, val := range u.Sets {
		ctyVal, err := hcltools.InterfaceToCty(val)
		if err != nil {
			return err
		}
		setBlock := helmBody.AppendNewBlock("set", []string{})
		setBlock.Body().SetAttributeValue("name", cty.StringVal(key))
		setBlock.Body().SetAttributeValue("value", ctyVal)
	}
	return nil
}

// generateV3SetValues generates set values for Helm provider v3.x+ (list of objects)
func (u *Unit) generateV3SetValues(helmBody *hclwrite.Body) error {
	setList := make([]cty.Value, 0, len(u.Sets))
	
	for key, val := range u.Sets {
		ctyVal, err := hcltools.InterfaceToCty(val)
		if err != nil {
			return err
		}
		
		// Create set object with name and value
		setObj := cty.ObjectVal(map[string]cty.Value{
			"name":  cty.StringVal(key),
			"value": ctyVal,
		})
		setList = append(setList, setObj)
	}
	
	// Set as list attribute
	helmBody.SetAttributeValue("set", cty.ListVal(setList))
	return nil
}

func (u *Unit) ReadConfig(spec map[string]interface{}, stack *project.Stack) error {
	err := utils.YAMLInterfaceToType(spec, u)
	if err != nil {
		return fmt.Errorf("read 'provider_conf': %w", err)
	}
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
				setData, ok := valuesCatMap["set"].(map[string]interface{})
				if !ok {
					return fmt.Errorf("read unit config: one of 'values.file' or 'values.set' required")
				}
				u.ValuesYAML = append(u.ValuesYAML, setData)
				yamlDaya, err := yaml.Marshal(setData)
				if err != nil {
					return fmt.Errorf("read unit config: %w", err)
				}
				u.ValuesFilesList = append(u.ValuesFilesList, string(yamlDaya))
				continue
			}
			vfPath := filepath.Join(u.Stack().TemplateDir, valuesFileName)
			valuesFileContent, err := os.ReadFile(vfPath)
			if err != nil {
				log.Debugf(err.Error())
				return fmt.Errorf("read unit config: can't load values file: %v", err)
			}
			values := valuesFileContent
			if applyTemplate {
				renderedValues, errIsWarn, err := u.Stack().TemplateTry(valuesFileContent, vfPath)
				if err != nil {
					if !errIsWarn {
						return err
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
		u.ProviderVersion = pv
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
	if err = u.CreateFiles.AddOverride("main.tf", string(mainBlock), fs.ModePerm); err != nil {
		return err
	}
	return u.Unit.Build()

}

// UpdateProjectRuntimeData update project runtime dataset, adds unit outputs.
func (u *Unit) UpdateProjectRuntimeData(p *project.Project) error {
	return u.Unit.UpdateProjectRuntimeData(p)
}
