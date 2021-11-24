package helm

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"path/filepath"
	"reflect"

	"github.com/apex/log"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/shalb/cluster.dev/pkg/hcltools"
	"github.com/shalb/cluster.dev/pkg/modules/shell/terraform/base"
	"github.com/shalb/cluster.dev/pkg/project"
	"github.com/shalb/cluster.dev/pkg/utils"
	"github.com/zclconf/go-cty/cty"
	"gopkg.in/yaml.v3"
)

type Unit struct {
	base.Unit
	StatePtr        *Unit                    `yaml:"-" json:"-"`
	Source          string                   `yaml:"-" json:"source"`
	HelmOpts        map[string]interface{}   `yaml:"-" json:"helm_opts,omitempty"`
	Sets            map[string]interface{}   `yaml:"-" json:"sets,omitempty"`
	Kubeconfig      string                   `yaml:"-" json:"kubeconfig"`
	ValuesFilesList []string                 `yaml:"-" json:"values,omitempty"`
	ValuesYAML      []map[string]interface{} `yaml:"-" json:"-"`
	UnitKind        string                   `yaml:"-" json:"type"`
}

func (m *Unit) KindKey() string {
	return unitKind
}

func (m *Unit) genMainCodeBlock() ([]byte, error) {
	// var marker string
	// if len(m.valuesFileContent) > 0 {
	// 	m.FilesList()["values.yaml"] = m.valuesFileContent
	// }
	f := hclwrite.NewEmptyFile()
	rootBody := f.Body()
	providerBlock := rootBody.AppendNewBlock("provider", []string{"helm"})
	providerBody := providerBlock.Body()
	provederKubernetesBlock := providerBody.AppendNewBlock("kubernetes", []string{})
	provederKubernetesBlock.Body().SetAttributeValue("config_path", cty.StringVal(m.Kubeconfig))

	helmBlock := rootBody.AppendNewBlock("resource", []string{"helm_release", project.ConvertToTfVarName(m.Name())})
	helmBody := helmBlock.Body()
	for key, val := range m.HelmOpts {
		ctyVal, err := hcltools.InterfaceToCty(val)
		if err != nil {
			return nil, err
		}
		helmBody.SetAttributeValue(key, ctyVal)
	}
	for key, val := range m.Sets {
		ctyVal, err := hcltools.InterfaceToCty(val)
		if err != nil {
			return nil, err
		}
		setBlock := helmBody.AppendNewBlock("set", []string{})
		setBlock.Body().SetAttributeValue("name", cty.StringVal(key))
		setBlock.Body().SetAttributeValue("value", ctyVal)
	}
	if len(m.ValuesFilesList) > 0 {
		ctyValuesList := []cty.Value{}
		for _, v := range m.ValuesFilesList {
			ctyValuesList = append(ctyValuesList, cty.StringVal(string(v)))
		}
		helmBody.SetAttributeValue("values", cty.ListVal(ctyValuesList))
		//hcltools.ReplaceStingMarkerInBody(helmBody, marker, "file(\"./values.yaml\")")
	}
	markersList := map[string]*project.DependencyOutput{}
	err := m.Project().GetMarkers(base.RemoteStateMarkerCatName, &markersList)
	if err != nil {
		return nil, err
	}
	for hash, marker := range markersList {
		if marker.StackName == "this" {
			marker.StackName = m.Stack().Name
		}
		refStr := base.DependencyToRemoteStateRef(marker)
		hcltools.ReplaceStingMarkerInBody(helmBody, hash, refStr)
	}
	return f.Bytes(), nil
}

func (m *Unit) ReadConfig(spec map[string]interface{}, stack *project.Stack) error {

	source, ok := spec["source"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("read unit config: incorrect unit source, %v", m.Key())
	}
	for key, val := range source {
		m.HelmOpts[key] = val
	}
	kubeconfig, ok := spec["kubeconfig"].(string)
	if !ok {
		return fmt.Errorf("Incorrect kubeconfig")
	}
	m.Kubeconfig = kubeconfig
	addOp, ok := spec["additional_options"].(map[string]interface{})
	if ok {
		for key, val := range addOp {
			m.HelmOpts[key] = val
		}
	}
	sets, ok := spec["inputs"].(map[string]interface{})
	if ok {
		for key, val := range sets {
			m.Sets[key] = val
		}
	}
	m.HelmOpts["name"] = spec["name"]
	valuesCat, ok := spec["values"]
	if ok {
		valuesCatList, check := valuesCat.([]interface{})
		if !check {
			return fmt.Errorf("read unit config: 'values' have unknown type: %v", reflect.TypeOf(valuesCat))
		}
		m.ValuesFilesList = []string{}
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
			vfPath := filepath.Join(m.Stack().TemplateDir, valuesFileName)
			valuesFileContent, err := ioutil.ReadFile(vfPath)
			if err != nil {
				log.Debugf(err.Error())
				return fmt.Errorf("read unit config: can't load values file: %v", err)
			}
			values := valuesFileContent
			if applyTemplate {
				renderedValues, errIsWarn, err := m.Stack().TemplateTry(valuesFileContent)
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
				return fmt.Errorf("read unit config: unmarshal values file: ", utils.ResolveYamlError(values, err))
			}
			m.ValuesYAML = append(m.ValuesYAML, vYAML)
			m.ValuesFilesList = append(m.ValuesFilesList, string(values))
		}
	}
	pv, ok := spec["provider_version"].(string)
	if ok {
		m.AddRequiredProvider("helm", "hashicorp/helm", pv)
	}
	m.StatePtr = &Unit{
		Unit: m.Unit,
	}
	err := utils.JSONCopy(m, m.StatePtr)
	return err
}

// ReplaceMarkers replace all templated markers with values.
func (m *Unit) ReplaceMarkers() error {
	err := m.Unit.ReplaceMarkers()
	if err != nil {
		return err
	}
	err = project.ScanMarkers(m.ValuesFilesList, m.RemoteStatesScanner, m)
	if err != nil {
		return err
	}
	err = project.ScanMarkers(m.HelmOpts, m.RemoteStatesScanner, m)
	if err != nil {
		return err
	}
	err = project.ScanMarkers(m.Sets, m.RemoteStatesScanner, m)
	if err != nil {
		return err
	}
	err = project.ScanMarkers(m.ValuesYAML, m.RemoteStatesScanner, m)
	if err != nil {
		return err
	}

	err = project.ScanMarkers(m.ValuesFilesList, project.OutputsScanner, m)
	if err != nil {
		return err
	}
	err = project.ScanMarkers(m.HelmOpts, project.OutputsScanner, m)
	if err != nil {
		return err
	}
	err = project.ScanMarkers(m.Sets, project.OutputsScanner, m)
	if err != nil {
		return err
	}
	err = project.ScanMarkers(m.ValuesYAML, project.OutputsScanner, m)
	if err != nil {
		return err
	}
	return nil
}

// Build generate all terraform code for project.
func (m *Unit) Build() error {

	mainBlock, err := m.genMainCodeBlock()
	if err != nil {
		log.Debug(err.Error())
		return err
	}
	if err = m.CreateFiles.Add("main.tf", string(mainBlock), fs.ModePerm); err != nil {
		return err
	}
	return m.Unit.Build()

}

// UpdateProjectRuntimeData update project runtime dataset, adds unit outputs.
func (m *Unit) UpdateProjectRuntimeData(p *project.Project) error {
	return m.Unit.UpdateProjectRuntimeData(p)
}
