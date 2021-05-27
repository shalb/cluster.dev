package kubernetes

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/apex/log"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/shalb/cluster.dev/pkg/hcltools"
	"github.com/shalb/cluster.dev/pkg/modules/terraform/common"
	"github.com/shalb/cluster.dev/pkg/project"
	"github.com/shalb/cluster.dev/pkg/utils"
)

type Module struct {
	common.Module
	source          string
	kubeconfig      string
	inputs          map[string]interface{}
	providerVersion string
	ProviderConf    ProviderConfigSpec
}

type ExecNestedSchema struct {
	APIVersion string            `yaml:"api_version,omitempty" json:"api_version,omitempty"`
	Args       []string          `yaml:"args,omitempty" json:"args,omitempty"`
	Command    string            `yaml:"command,omitempty" json:"command,omitempty"`
	Env        map[string]string `yaml:"env,omitempty" json:"env,omitempty"`
}

type ProviderConfigSpec struct {
	ConfigPath           string            `yaml:"config_path,omitempty" json:"config_path,omitempty"`
	ClientCertificate    string            `yaml:"client_certificate,omitempty" json:"client_certificate,omitempty"`
	ConfigContext        string            `yaml:"config_context,omitempty" json:"config_context,omitempty"`
	ConfigContextCluster string            `yaml:"config_context_cluster,omitempty" json:"config_context_cluster,omitempty"`
	ConfigContextUser    string            `yaml:"config_context_user,omitempty"  json:"config_context_user,omitempty"`
	Exec                 *ExecNestedSchema `yaml:"exec,omitempty" json:"exec,omitempty"`
	Host                 string            `yaml:"host,omitempty" json:"host,omitempty"`
	Insecure             string            `yaml:"insecure,omitempty" json:"insecure,omitempty"`
	Password             string            `yaml:"password,omitempty" json:"password,omitempty"`
	Token                string            `yaml:"token,omitempty" json:"token,omitempty"`
	Username             string            `yaml:"username,omitempty" json:"username,omitempty"`
}

func (m *Module) KindKey() string {
	return "kubernetes"
}

func (m *Module) genMainCodeBlock() ([]byte, error) {
	f := hclwrite.NewEmptyFile()
	rootBody := f.Body()
	providerBlock := rootBody.AppendNewBlock("provider", []string{"kubernetes-alpha"})
	providerBody := providerBlock.Body()

	providerCty, err := hcltools.InterfaceToCty(m.ProviderConf)
	if err != nil {
		return nil, err
	}
	for key, val := range providerCty.AsValueMap() {
		providerBody.SetAttributeValue(key, val)
	}
	for key, manifest := range m.inputs {
		moduleBlock := rootBody.AppendNewBlock("resource", []string{"kubernetes_manifest", key})
		moduleBody := moduleBlock.Body()
		tokens := hclwrite.Tokens{&hclwrite.Token{Type: hclsyntax.TokenQuotedLit, Bytes: []byte(" kubernetes-alpha"), SpacesBefore: 1}}
		moduleBody.SetAttributeRaw("provider", tokens)
		ctyVal, err := hcltools.InterfaceToCty(manifest)
		if err != nil {
			return nil, err
		}

		moduleBody.SetAttributeValue("manifest", ctyVal)
		depMarkers, ok := m.ProjectPtr().Markers[common.RemoteStateMarkerCatName]
		if ok {
			for hash, marker := range depMarkers.(map[string]*project.Dependency) {
				if marker.Module == nil {
					continue
				}
				remoteStateRef := fmt.Sprintf("data.terraform_remote_state.%s-%s.outputs.%s", marker.Module.InfraName(), marker.Module.Name(), marker.Output)
				hcltools.ReplaceStingMarkerInBody(moduleBody, hash, remoteStateRef)
			}
		}
	}
	depMarkers, ok := m.ProjectPtr().Markers[common.RemoteStateMarkerCatName]
	if ok {
		for hash, marker := range depMarkers.(map[string]*project.Dependency) {
			if marker.Module == nil {
				continue
			}
			remoteStateRef := fmt.Sprintf("data.terraform_remote_state.%s-%s.outputs.%s", marker.Module.InfraName(), marker.Module.Name(), marker.Output)
			hcltools.ReplaceStingMarkerInBody(providerBody, hash, remoteStateRef)
		}
	}

	return f.Bytes(), nil
}

func (m *Module) ReadConfig(spec map[string]interface{}, infra *project.Infrastructure) error {
	err := m.ReadConfigCommon(spec, infra)
	if err != nil {
		return fmt.Errorf("reading kubernetes module: %v", err.Error())
	}
	source, ok := spec["source"].(string)
	if !ok {
		return fmt.Errorf("reading kubernetes module '%v': malformed module source", m.Key())
	}
	tmplDir := m.InfraPtr().TemplateDir
	var absSource string
	if source[1:2] == "/" {
		absSource = filepath.Join(tmplDir, source)
	} else {
		absSource = source
	}
	fileInfo, err := os.Stat(absSource)
	if err != nil {
		return fmt.Errorf("reading kubernetes module '%v': reading kubernetes manifests form source '%v': %v", m.Key(), source, err.Error())
	}
	var filesList []string
	if fileInfo.IsDir() {
		filesList, err = filepath.Glob(absSource + "/*.yaml")
		if err != nil {
			return fmt.Errorf("reading kubernetes module '%v': reading kubernetes manifests form source '%v': %v", m.Key(), source, err.Error())
		}
	} else {
		filesList = append(filesList, absSource)
	}
	for _, fileName := range filesList {
		file, err := ioutil.ReadFile(fileName)
		if err != nil {
			return fmt.Errorf("reading kubernetes module '%v': reading kubernetes manifests form source '%v': %v", m.Key(), source, err.Error())
		}
		manifest, errIsWarn, err := m.InfraPtr().TemplateTry(file)
		if err != nil {
			if errIsWarn {
				log.Warnf("File %v has unresolved template key: \n%v", fileName, err.Error())
			} else {
				log.Fatal(err.Error())
			}
		}
		manifests, err := utils.ReadYAMLObjects(manifest)
		if err != nil {
			return fmt.Errorf("reading kubernetes module '%v': reading kubernetes manifests form source '%v': %v", m.Key(), source, err.Error())
		}

		for i, manifest := range manifests {
			key := project.ConvertToTfVarName(strings.TrimSuffix(filepath.Base(fileName), ".yaml"))
			key = fmt.Sprintf("%s_%v", key, i)
			m.inputs[key] = manifest
		}
	}
	if len(m.inputs) < 1 {
		return fmt.Errorf("the kubernetes module must contain at least one manifest")
	}

	err = utils.JSONInterfaceToType(spec, &m.ProviderConf)
	if err != nil {
		return err
	}
	kubeconfig, ok := spec["kubeconfig"].(string)
	if ok && m.ProviderConf.ConfigPath == "" {
		m.ProviderConf.ConfigPath = kubeconfig
	}
	pv, ok := spec["provider_version"].(string)
	if ok {
		m.AddRequiredProvider("kubernetes-alpha", "hashicorp/kubernetes-alpha", pv)
	}
	m.source = source
	return nil
}

// ReplaceMarkers replace all templated markers with values.
func (m *Module) ReplaceMarkers() error {
	err := m.ReplaceMarkersCommon(m)
	if err != nil {
		return err
	}
	err = project.ScanMarkers(m.inputs, m.YamlBlockMarkerScanner, m)
	if err != nil {
		return err
	}
	err = project.ScanMarkers(m.inputs, m.RemoteStatesScanner, m)
	if err != nil {
		return err
	}
	err = project.ScanMarkers(&m.ProviderConf, m.RemoteStatesScanner, m)
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
	return m.CreateCodeDir()
}

// UpdateProjectRuntimeData update project runtime dataset, adds module outputs.
func (m *Module) UpdateProjectRuntimeData(p *project.Project) error {
	return m.UpdateProjectRuntimeDataCommon(p)
}
