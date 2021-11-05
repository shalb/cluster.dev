package helm

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"path/filepath"

	"github.com/apex/log"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/shalb/cluster.dev/pkg/hcltools"
	"github.com/shalb/cluster.dev/pkg/modules/shell/terraform/base"
	"github.com/shalb/cluster.dev/pkg/project"
	"github.com/zclconf/go-cty/cty"
	"gopkg.in/yaml.v3"
)

type Unit struct {
	base.Unit
	source          string
	helmOpts        map[string]interface{}
	sets            map[string]interface{}
	kubeconfig      string
	valuesFilesList []string
	valuesYAML      []map[string]interface{}
}

func (m *Unit) KindKey() string {
	return "helm"
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
	provederKubernetesBlock.Body().SetAttributeValue("config_path", cty.StringVal(m.kubeconfig))

	helmBlock := rootBody.AppendNewBlock("resource", []string{"helm_release", project.ConvertToTfVarName(m.Name())})
	helmBody := helmBlock.Body()
	for key, val := range m.helmOpts {
		ctyVal, err := hcltools.InterfaceToCty(val)
		if err != nil {
			return nil, err
		}
		helmBody.SetAttributeValue(key, ctyVal)
	}
	for key, val := range m.sets {
		ctyVal, err := hcltools.InterfaceToCty(val)
		if err != nil {
			return nil, err
		}
		setBlock := helmBody.AppendNewBlock("set", []string{})
		setBlock.Body().SetAttributeValue("name", cty.StringVal(key))
		setBlock.Body().SetAttributeValue("value", ctyVal)
	}
	if len(m.valuesFilesList) > 0 {
		ctyValuesList := []cty.Value{}
		for _, v := range m.valuesFilesList {
			ctyValuesList = append(ctyValuesList, cty.StringVal(string(v)))
		}
		helmBody.SetAttributeValue("values", cty.ListVal(ctyValuesList))
		//hcltools.ReplaceStingMarkerInBody(helmBody, marker, "file(\"./values.yaml\")")
	}
	for hash, m := range m.Markers() {
		marker, ok := m.(*project.DependencyOutput)
		// log.Warnf("kubernetes marker HELM: %v", marker)
		refStr := base.DependencyToRemoteStateRef(marker)
		if !ok {
			return nil, fmt.Errorf("generate main.tf: internal error: incorrect remote state type")
		}
		hcltools.ReplaceStingMarkerInBody(helmBody, hash, refStr)
	}
	return f.Bytes(), nil
}

func (m *Unit) ReadConfig(spec map[string]interface{}, stack *project.Stack) error {
	err := m.Unit.ReadConfig(spec, stack)
	if err != nil {
		log.Debug(err.Error())
		return err
	}
	source, ok := spec["source"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("read unit config: incorrect unit source, %v", m.Key())
	}
	for key, val := range source {
		m.helmOpts[key] = val
	}
	kubeconfig, ok := spec["kubeconfig"].(string)
	if !ok {
		return fmt.Errorf("Incorrect kubeconfig")
	}
	m.kubeconfig = kubeconfig
	addOp, ok := spec["additional_options"].(map[string]interface{})
	if ok {
		for key, val := range addOp {
			m.helmOpts[key] = val
		}
	}
	sets, ok := spec["inputs"].(map[string]interface{})
	if ok {
		for key, val := range sets {
			m.sets[key] = val
		}
	}
	m.helmOpts["name"] = spec["name"]
	valuesCat, ok := spec["values"]
	if ok {
		valuesCatList, check := valuesCat.([]interface{})
		if !check {
			return fmt.Errorf("read unit config: 'values' have unknown type: %v", err)
		}
		m.valuesFilesList = []string{}
		//log.Warnf("%v", ok)

		for _, valuesCat := range valuesCatList {
			valuesCatMap, check := valuesCat.(map[string]interface{})
			if !check {
				return fmt.Errorf("read unit config: 'values' have unknown format: %v", err)
			}
			applyTemplate, exists := valuesCatMap["apply_template"].(bool)
			if !exists {
				applyTemplate = true
			}
			valuesFileName, ok := valuesCatMap["file"].(string)
			if !ok {
				return fmt.Errorf("read unit config: 'values.file' is required field: %v", err)
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
				return fmt.Errorf("read unit config: unmarshal values file: ", err.Error())
			}
			m.valuesYAML = append(m.valuesYAML, vYAML)
			m.valuesFilesList = append(m.valuesFilesList, string(values))
		}
	}
	pv, ok := spec["provider_version"].(string)
	if ok {
		m.AddRequiredProvider("helm", "hashicorp/helm", pv)
	}
	return nil
}

// ReplaceMarkers replace all templated markers with values.
func (m *Unit) ReplaceMarkers() error {
	err := m.Unit.ReplaceMarkers(m)
	if err != nil {
		return err
	}
	err = project.ScanMarkers(m.valuesFilesList, m.RemoteStatesScanner, m)
	if err != nil {
		return err
	}
	err = project.ScanMarkers(m.helmOpts, m.RemoteStatesScanner, m)
	if err != nil {
		return err
	}
	err = project.ScanMarkers(m.sets, m.RemoteStatesScanner, m)
	if err != nil {
		return err
	}
	err = project.ScanMarkers(m.valuesYAML, m.RemoteStatesScanner, m)
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
