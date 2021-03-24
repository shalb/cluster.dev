package s3

import (
	"fmt"

	"github.com/zclconf/go-cty/cty"

	"github.com/hashicorp/hcl/v2/hclwrite"
)

// BackendS3 - describe s3 backend for interface package.backend.
type BackendS3 struct {
	name   string
	Bucket string `yaml:"bucket"`
	Region string `yaml:"region"`
	state  map[string]interface{}
}

func (b *BackendS3) State() map[string]interface{} {
	return b.state
}

// Name return name.
func (b *BackendS3) Name() string {
	return b.name
}

// Provider return name.
func (b *BackendS3) Provider() string {
	return "s3"
}

type backendConfigSpec struct {
	Bucket string `hcl:"bucket"`
	Key    string `hcl:"key"`
	Region string `hcl:"region"`
}

// GetBackendBytes generate terraform backend config.
func (b *BackendS3) GetBackendBytes(infraName, moduleName string) ([]byte, error) {
	f, err := b.GetBackendHCL(infraName, moduleName)
	if err != nil {
		return nil, err
	}
	return f.Bytes(), nil
}

// GetBackendHCL generate terraform backend config.
func (b *BackendS3) GetBackendHCL(infraName, moduleName string) (*hclwrite.File, error) {
	f := hclwrite.NewEmptyFile()
	rootBody := f.Body()
	terraformBlock := rootBody.AppendNewBlock("terraform", []string{})
	backendBlock := terraformBlock.Body().AppendNewBlock("backend", []string{"s3"})
	backendBody := backendBlock.Body()
	backendBody.SetAttributeValue("bucket", cty.StringVal(b.Bucket))
	backendBody.SetAttributeValue("key", cty.StringVal(fmt.Sprintf("%s/%s.state", infraName, moduleName)))
	backendBody.SetAttributeValue("region", cty.StringVal(b.Region))

	terraformBlock.Body().SetAttributeValue("required_version", cty.StringVal("~> 0.13"))
	return f, nil

}

// GetRemoteStateHCL generate terraform remote state for this backend.
func (b *BackendS3) GetRemoteStateHCL(infraName, moduleName string) ([]byte, error) {
	f := hclwrite.NewEmptyFile()

	rootBody := f.Body()
	dataBlock := rootBody.AppendNewBlock("data", []string{"terraform_remote_state", fmt.Sprintf("%s-%s", infraName, moduleName)})
	dataBody := dataBlock.Body()
	dataBody.SetAttributeValue("backend", cty.StringVal("s3"))
	dataBody.SetAttributeValue("config", cty.MapVal(map[string]cty.Value{
		"bucket": cty.StringVal(b.Bucket),
		"key":    cty.StringVal(fmt.Sprintf("%s/%s.state", infraName, moduleName)),
		"region": cty.StringVal(b.Region),
	}))

	return f.Bytes(), nil
}
