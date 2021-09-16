package helm

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/apex/log"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/shalb/cluster.dev/pkg/hcltools"
	"github.com/shalb/cluster.dev/pkg/modules/terraform/common"
	"github.com/shalb/cluster.dev/pkg/project"
	"github.com/zclconf/go-cty/cty"
	"gopkg.in/yaml.v3"
)

type Module struct {
	common.Module
	source          string
	helmOpts        map[string]interface{}
	sets            map[string]interface{}
	kubeconfig      string
	valuesFilesList []string
	valuesYAML      []map[string]interface{}
}

func (m *Module) KindKey() string {
	return "helm"
}

func (m *Module) genMainCodeBlock() ([]byte, error) {
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
		refStr := common.DependencyToRemoteStateRef(marker)
		if !ok {
			return nil, fmt.Errorf("generate main.tf: internal error: incorrect remote state type")
		}
		hcltools.ReplaceStingMarkerInBody(helmBody, hash, refStr)
	}
	return f.Bytes(), nil
}

func (m *Module) ReadConfig(spec map[string]interface{}, infra *project.Infrastructure) error {
	err := m.ReadConfigCommon(spec, infra)
	if err != nil {
		log.Debug(err.Error())
		return err
	}
	source, ok := spec["source"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("read module config: incorrect module source, %v", m.Key())
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
			return fmt.Errorf("read module config: 'values' have unknown type: %v", err)
		}
		m.valuesFilesList = []string{}
		//log.Warnf("%v", ok)

		for _, valuesCat := range valuesCatList {
			valuesCatMap, check := valuesCat.(map[string]interface{})
			if !check {
				return fmt.Errorf("read module config: 'values' have unknown format: %v", err)
			}
			applyTemplate, exists := valuesCatMap["apply_template"].(bool)
			if !exists {
				applyTemplate = true
			}
			valuesFileName, ok := valuesCatMap["file"].(string)
			if !ok {
				return fmt.Errorf("read module config: 'values.file' is required field: %v", err)
			}
			vfPath := filepath.Join(m.InfraPtr().TemplateDir, valuesFileName)
			valuesFileContent, err := ioutil.ReadFile(vfPath)
			if err != nil {
				log.Debugf(err.Error())
				return fmt.Errorf("read module config: can't load values file: %v", err)
			}
			values := valuesFileContent
			if applyTemplate {
				renderedValues, errIsWarn, err := m.InfraPtr().TemplateTry(valuesFileContent)
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
				return fmt.Errorf("read module config: unmarshal values file: ", err.Error())
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
func (m *Module) ReplaceMarkers() error {
	err := m.ReplaceMarkersCommon(m)
	if err != nil {
		return err
	}
	err = project.ScanMarkers(m.helmOpts, project.YamlBlockMarkerScanner, m)
	if err != nil {
		return err
	}
	err = project.ScanMarkers(m.valuesFilesList, m.RemoteStatesScanner, m)
	if err != nil {
		return err
	}
	err = project.ScanMarkers(m.valuesFilesList, project.YamlBlockMarkerScanner, m)
	if err != nil {
		return err
	}
	err = project.ScanMarkers(m.helmOpts, m.RemoteStatesScanner, m)
	if err != nil {
		return err
	}
	err = project.ScanMarkers(m.sets, project.YamlBlockMarkerScanner, m)
	if err != nil {
		return err
	}
	err = project.ScanMarkers(m.sets, m.RemoteStatesScanner, m)
	if err != nil {
		return err
	}
	err = project.ScanMarkers(m.valuesYAML, project.YamlBlockMarkerScanner, m)
	if err != nil {
		return err
	}
	err = project.ScanMarkers(m.valuesYAML, m.RemoteStatesScanner, m)
	if err != nil {
		return err
	}
	return nil
}

// CreateCodeDir generate all terraform code for project.
func (m *Module) Build() error {
	m.BuildCommon()
	var err error
	m.FilesList()["main.tf"], err = m.genMainCodeBlock()
	if err != nil {
		log.Debug(err.Error())
		return err
	}
	//	log.Debugf("VALUES: %v", string(m.valuesFileContent))
	return m.CreateCodeDir()
}

// UpdateProjectRuntimeData update project runtime dataset, adds module outputs.
func (m *Module) UpdateProjectRuntimeData(p *project.Project) error {
	return m.UpdateProjectRuntimeDataCommon(p)
}
