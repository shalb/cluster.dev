package azurerm

import (
	"fmt"

	"github.com/shalb/cluster.dev/pkg/hcltools"
	"github.com/zclconf/go-cty/cty"

	"github.com/hashicorp/hcl/v2/hclwrite"
)

// BackendAzureRm - describe s3 backend for interface package.backend.
type BackendAzureRm struct {
	name  string
	state map[string]interface{}
}

func (b *BackendAzureRm) State() map[string]interface{} {
	return b.state
}

// Name return name.
func (b *BackendAzureRm) Name() string {
	return b.name
}

// Provider return name.
func (b *BackendAzureRm) Provider() string {
	return "azurerm"
}

// GetBackendBytes generate terraform backend config.
func (b *BackendAzureRm) GetBackendBytes(infraName, moduleName string) ([]byte, error) {
	f, err := b.GetBackendHCL(infraName, moduleName)
	if err != nil {
		return nil, err
	}
	return f.Bytes(), nil
}

// GetBackendHCL generate terraform backend config.
func (b *BackendAzureRm) GetBackendHCL(infraName, moduleName string) (*hclwrite.File, error) {
	b.state["key"] = fmt.Sprintf("%s-%s.state", infraName, moduleName)

	f := hclwrite.NewEmptyFile()
	rootBody := f.Body()
	terraformBlock := rootBody.AppendNewBlock("terraform", []string{})
	backendBlock := terraformBlock.Body().AppendNewBlock("backend", []string{"azurerm"})
	backendBody := backendBlock.Body()
	for key, value := range b.state {
		backendBody.SetAttributeValue(key, cty.StringVal(value.(string)))
	}
	return f, nil
}

// GetRemoteStateHCL generate terraform remote state for this backend.
func (b *BackendAzureRm) GetRemoteStateHCL(infraName, moduleName string) ([]byte, error) {
	b.state["key"] = fmt.Sprintf("%s-%s.state", infraName, moduleName)

	f := hclwrite.NewEmptyFile()

	rootBody := f.Body()
	dataBlock := rootBody.AppendNewBlock("data", []string{"terraform_remote_state", fmt.Sprintf("%s-%s", infraName, moduleName)})
	dataBody := dataBlock.Body()
	dataBody.SetAttributeValue("backend", cty.StringVal("azurerm"))
	config, err := hcltools.InterfaceToCty(b.state)
	if err != nil {
		return nil, err
	}
	dataBody.SetAttributeValue("config", config)
	return f.Bytes(), nil
}
