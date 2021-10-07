package gcs

import (
	"fmt"

	"github.com/shalb/cluster.dev/pkg/hcltools"
	"github.com/shalb/cluster.dev/pkg/project"
	"github.com/zclconf/go-cty/cty"
	"gopkg.in/yaml.v3"

	"github.com/hashicorp/hcl/v2/hclwrite"
)

// Backend - describe s3 backend for interface package.backend.
type Backend struct {
	name                      string
	Bucket                    string `yaml:"bucket"`
	Credentials               string `yaml:"credentials,omitempty"`
	ImpersonateServiceAccount string `yaml:"impersonate_service_account,omitempty"`
	AccessToken               string `yaml:"access_token,omitempty"`
	EncryptionKey             string `yaml:"encryption_key,omitempty"`
	Prefix                    string `yaml:"prefix"`
	state                     map[string]interface{}
	ProjectPtr                *project.Project
}

func (b *Backend) State() map[string]interface{} {
	return b.state
}

// Name return name.
func (b *Backend) Name() string {
	return b.name
}

// Provider return name.
func (b *Backend) Provider() string {
	return "gcs"
}

// GetBackendBytes generate terraform backend config.
func (b *Backend) GetBackendBytes(stackName, unitName string) ([]byte, error) {
	f, err := b.GetBackendHCL(stackName, unitName)
	if err != nil {
		return nil, err
	}
	return f.Bytes(), nil
}

// GetBackendHCL generate terraform backend config.
func (b *Backend) GetBackendHCL(stackName, unitName string) (*hclwrite.File, error) {
	bConfigTmpl, err := getStateMap(*b)
	if err != nil {
		return nil, err
	}
	bConfigTmpl["prefix"] = fmt.Sprintf("%s%s_%s", b.Prefix, stackName, unitName)
	f := hclwrite.NewEmptyFile()
	rootBody := f.Body()
	terraformBlock := rootBody.AppendNewBlock("terraform", []string{})
	backendBlock := terraformBlock.Body().AppendNewBlock("backend", []string{"gcs"})
	backendBody := backendBlock.Body()
	for key, value := range bConfigTmpl {
		backendBody.SetAttributeValue(key, cty.StringVal(value.(string)))
	}
	return f, nil
}

// GetRemoteStateHCL generate terraform remote state for this backend.
func (b *Backend) GetRemoteStateHCL(stackName, unitName string) ([]byte, error) {
	bConfigTmpl, err := getStateMap(*b)
	if err != nil {
		return nil, err
	}
	bConfigTmpl["prefix"] = fmt.Sprintf("%s%s_%s", b.Prefix, stackName, unitName)
	f := hclwrite.NewEmptyFile()

	rootBody := f.Body()
	dataBlock := rootBody.AppendNewBlock("data", []string{"terraform_remote_state", fmt.Sprintf("%s-%s", stackName, unitName)})
	dataBody := dataBlock.Body()
	dataBody.SetAttributeValue("backend", cty.StringVal("gcs"))
	config, err := hcltools.InterfaceToCty(bConfigTmpl)
	if err != nil {
		return nil, err
	}
	dataBody.SetAttributeValue("config", config)
	return f.Bytes(), nil
}

func getStateMap(in Backend) (res map[string]interface{}, err error) {
	tmpData, err := yaml.Marshal(in)
	if err != nil {
		return
	}
	res = map[string]interface{}{}
	err = yaml.Unmarshal(tmpData, &res)
	return
}

func (b *Backend) LockState() error {
	return fmt.Errorf("cdev state gcs not supported")
}

func (b *Backend) UnlockState() error {
	return fmt.Errorf("cdev state gcs not supported")
}

func (b *Backend) WriteState(stateData string) error {
	return fmt.Errorf("cdev state gcs not supported")
}

func (b *Backend) ReadState() (string, error) {
	return "", fmt.Errorf("cdev state gcs not supported")
}
