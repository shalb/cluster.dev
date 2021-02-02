package helm

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"reflect"

	"github.com/apex/log"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/shalb/cluster.dev/pkg/hcltools"
	"github.com/shalb/cluster.dev/pkg/modules/terraform/common"
	"github.com/shalb/cluster.dev/pkg/project"
	"github.com/zclconf/go-cty/cty"
)

type helm struct {
	common.Module
	source            string
	helmOpts          map[string]interface{}
	sets              map[string]interface{}
	kubeconfig        string
	valuesFileContent []byte
}

func (m *helm) genMainCodeBlock() ([]byte, error) {
	var marker string
	if len(m.valuesFileContent) > 0 {
		m.FilesList["values.yaml"] = m.valuesFileContent
		marker = project.CreateMarker("func")
		m.helmOpts["values"] = []string{string(marker)}
	}
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
	depMarkers, ok := m.ProjectPtr().Markers[common.RemoteStateMarkerCatName]
	if ok {
		for hash, marker := range depMarkers.(map[string]*project.Dependency) {
			if marker.Module == nil {
				continue
			}
			remoteStateRef := fmt.Sprintf("data.terraform_remote_state.%s-%s.outputs.%s", marker.Module.InfraName(), marker.Module.Name(), marker.Output)
			hcltools.ReplaceStingMarkerInBody(helmBody, hash, remoteStateRef)
		}
	}
	if len(m.valuesFileContent) > 0 {
		hcltools.ReplaceStingMarkerInBody(helmBody, marker, "file(\"./values.yaml\")")
	}
	return f.Bytes(), nil
}

func (m *helm) ReadConfig(spec map[string]interface{}) error {
	source, ok := spec["source"].(map[string]interface{})
	log.Debugf("%v", reflect.TypeOf(spec["source"]))
	if !ok {
		return fmt.Errorf("Incorrect module source, %v", m.Key())
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
	valuesFile, ok := spec["values_file"].(string)
	if ok {
		vfPath := filepath.Join(filepath.Dir(m.InfraPtr().TemplateSrc), valuesFile)
		valuesFileContent, err := ioutil.ReadFile(vfPath)
		if err != nil {
			log.Debugf(err.Error())
			return fmt.Errorf("can't load values file: %v", err)
		}
		m.valuesFileContent = valuesFileContent
	}

	return nil
}

// ReplaceMarkers replace all templated markers with values.
func (m *helm) ReplaceMarkers() error {
	err := project.ScanMarkers(m.helmOpts, m.YamlBlockMarkerScanner, m)
	if err != nil {
		return err
	}
	err = project.ScanMarkers(m.helmOpts, m.RemoteStatesScanner, m)
	if err != nil {
		return err
	}
	err = project.ScanMarkers(m.sets, m.YamlBlockMarkerScanner, m)
	if err != nil {
		return err
	}
	err = project.ScanMarkers(m.sets, m.RemoteStatesScanner, m)
	if err != nil {
		return err
	}
	return nil
}

// CreateCodeDir generate all terraform code for project.
func (m *helm) Build(codeDir string) error {
	m.BuildCommon()
	var err error
	m.FilesList["main.tf"], err = m.genMainCodeBlock()
	if err != nil {
		log.Debug(err.Error())
		return err
	}
	//	log.Debugf("VALUES: %v", string(m.valuesFileContent))
	return m.CreateCodeDir(codeDir)
}
