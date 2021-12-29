package kubernetes

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/apex/log"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/shalb/cluster.dev/pkg/hcltools"
	"github.com/shalb/cluster.dev/pkg/units/shell/terraform/base"

	"github.com/shalb/cluster.dev/pkg/project"
	"github.com/shalb/cluster.dev/pkg/utils"
)

type Unit struct {
	base.Unit
	Source          string                 `yaml:"-" json:"source"`
	Kubeconfig      string                 `yaml:"-" json:"kubeconfig"`
	Inputs          map[string]interface{} `yaml:"-" json:"inputs"`
	providerVersion string                 `yaml:"-" json:"-"`
	ProviderConf    ProviderConfigSpec     `yaml:"-" json:"provider_conf"`
	UnitKind        string                 `yaml:"-" json:"type"`
	StateData       interface{}            `yaml:"-" json:"-"`
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

func (u *Unit) KindKey() string {
	return unitKind
}

func (u *Unit) genMainCodeBlock() ([]byte, error) {
	f := hclwrite.NewEmptyFile()
	rootBody := f.Body()
	providerBlock := rootBody.AppendNewBlock("provider", []string{"kubernetes-alpha"})
	providerBody := providerBlock.Body()

	providerCty, err := hcltools.InterfaceToCty(u.ProviderConf)
	if err != nil {
		return nil, err
	}
	for key, val := range providerCty.AsValueMap() {
		providerBody.SetAttributeValue(key, val)
	}
	for key, manifest := range u.Inputs {
		unitBlock := rootBody.AppendNewBlock("resource", []string{"kubernetes_manifest", key})
		unitBody := unitBlock.Body()
		tokens := hclwrite.Tokens{&hclwrite.Token{Type: hclsyntax.TokenQuotedLit, Bytes: []byte(" kubernetes-alpha"), SpacesBefore: 1}}
		unitBody.SetAttributeRaw("provider", tokens)
		ctyVal, err := hcltools.InterfaceToCty(manifest)
		if err != nil {
			return nil, err
		}

		unitBody.SetAttributeValue("manifest", ctyVal)
		for hash, marker := range u.ProjectPtr.UnitLinks.ByLinkTypes(base.RemoteStateLinkType).Map() {
			if marker.TargenStackName == "this" {
				marker.TargenStackName = u.Stack().Name
			}
			refStr := base.DependencyToRemoteStateRef(marker)
			hcltools.ReplaceStingMarkerInBody(unitBody, hash, refStr)
		}
	}

	for hash, marker := range u.ProjectPtr.UnitLinks.ByLinkTypes(base.RemoteStateLinkType).Map() {
		if marker.TargenStackName == "this" {
			marker.TargenStackName = u.Stack().Name
		}
		refStr := base.DependencyToRemoteStateRef(marker)
		hcltools.ReplaceStingMarkerInBody(providerBody, hash, refStr)
	}
	return f.Bytes(), nil
}

func (u *Unit) ReadConfig(spec map[string]interface{}, stack *project.Stack) error {
	source, ok := spec["source"].(string)
	if !ok {
		return fmt.Errorf("reading kubernetes unit '%v': malformed unit source", u.Key())
	}
	tmplDir := u.Stack().TemplateDir
	var absSource string
	if source[1:2] == "/" {
		absSource = filepath.Join(tmplDir, source)
	} else {
		absSource = source
	}
	fileInfo, err := os.Stat(absSource)
	if err != nil {
		return fmt.Errorf("reading kubernetes unit '%v': reading kubernetes manifests form source '%v': %v", u.Key(), source, err.Error())
	}
	var filesList []string
	if fileInfo.IsDir() {
		filesList, err = filepath.Glob(absSource + "/*.yaml")
		if err != nil {
			return fmt.Errorf("reading kubernetes unit '%v': reading kubernetes manifests form source '%v': %v", u.Key(), source, err.Error())
		}
	} else {
		filesList = append(filesList, absSource)
	}
	for _, fileName := range filesList {
		file, err := ioutil.ReadFile(fileName)
		if err != nil {
			return fmt.Errorf("reading kubernetes unit '%v': reading kubernetes manifests form source '%v': %v", u.Key(), source, err.Error())
		}
		manifest, errIsWarn, err := u.Stack().TemplateTry(file)
		if err != nil {
			if errIsWarn {
				log.Warnf("File %v has unresolved template key: \n%v", fileName, err.Error())
			} else {
				log.Fatal(err.Error())
			}
		}
		manifests, err := utils.ReadYAMLObjects(manifest)
		if err != nil {
			return fmt.Errorf("reading kubernetes unit '%v': reading kubernetes manifests form source '%v': %v", u.Key(), source, err.Error())
		}

		for i, manifest := range manifests {
			key := project.ConvertToTfVarName(strings.TrimSuffix(filepath.Base(fileName), ".yaml"))
			key = fmt.Sprintf("%s_%v", key, i)
			u.Inputs[key] = manifest
		}
	}
	if len(u.Inputs) < 1 {
		return fmt.Errorf("the kubernetes unit must contain at least one manifest")
	}

	err = utils.JSONCopy(spec, &u.ProviderConf)
	if err != nil {
		return err
	}
	kubeconfig, ok := spec["kubeconfig"].(string)
	if ok && u.ProviderConf.ConfigPath == "" {
		u.ProviderConf.ConfigPath = kubeconfig
	}
	pv, ok := spec["provider_version"].(string)
	if ok {
		u.AddRequiredProvider("kubernetes-alpha", "hashicorp/kubernetes-alpha", pv)
	}
	u.Source = source
	return nil
}

func (u *Unit) ScanData(scanner project.MarkerScanner) error {
	err := project.ScanMarkers(u.Inputs, scanner, u)
	if err != nil {
		return err
	}
	err = project.ScanMarkers(&u.ProviderConf, scanner, u)
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
	u.ScanData(project.OutputsReplacer)

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
