package local

import (
	"fmt"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/shalb/cluster.dev/internal/project"
	"github.com/zclconf/go-cty/cty"
)

// Backend - describe s3 backend for interface package.backend.
type Backend struct {
	name       string
	ProjectPtr *project.Project
	Path       string `yaml:"path"`
}

// Name return name.
func (b *Backend) Name() string {
	return b.name
}

// Provider return name.
func (b *Backend) Provider() string {
	return "local"
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
	f := hclwrite.NewEmptyFile()
	rootBody := f.Body()
	terraformBlock := rootBody.AppendNewBlock("terraform", []string{})
	backendBlock := terraformBlock.Body().AppendNewBlock("backend", []string{"local"})
	backendBody := backendBlock.Body()
	backendBody.SetAttributeValue("path", cty.StringVal(fmt.Sprintf("%s/%s.%s.tfstate", b.Path, stackName, unitName)))
	return f, nil

}

// GetRemoteStateHCL generate terraform remote state for this backend.
func (b *Backend) GetRemoteStateHCL(stackName, unitName string) ([]byte, error) {
	f := hclwrite.NewEmptyFile()

	rootBody := f.Body()
	dataBlock := rootBody.AppendNewBlock("data", []string{"terraform_remote_state", fmt.Sprintf("%s-%s", stackName, unitName)})
	dataBody := dataBlock.Body()
	dataBody.SetAttributeValue("backend", cty.StringVal("local"))
	dataBody.SetAttributeValue("config", cty.MapVal(map[string]cty.Value{
		"path": cty.StringVal(fmt.Sprintf("%s/%s.%s.tfstate", b.Path, stackName, unitName)),
	}))

	return f.Bytes(), nil
}
