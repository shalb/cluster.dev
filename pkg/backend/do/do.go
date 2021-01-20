package do

import (
	"fmt"

	"github.com/zclconf/go-cty/cty"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/shalb/cluster.dev/pkg/project"
)

// BackendS3 - describe s3 backend for interface package.backend.
type BackendDo struct {
	name   string
	Bucket string `yaml:"bucket"`
	Region string `yaml:"region"`
}

// Name return name.
func (b *BackendDo) Name() string {
	return b.name
}

// Provider return name.
func (b *BackendDo) Provider() string {
	return "s3"
}

type backendConfigSpec struct {
	Bucket string `hcl:"bucket"`
	Key    string `hcl:"key"`
	Region string `hcl:"region"`
}

// GetBackendHCL generate terraform backend config.
func (b *BackendDo) GetBackendHCL(module project.Module) ([]byte, error) {
	f := hclwrite.NewEmptyFile()
	rootBody := f.Body()
	terraformBlock := rootBody.AppendNewBlock("terraform", []string{})
	backendBlock := terraformBlock.Body().AppendNewBlock("backend", []string{"s3"})
	backendBody := backendBlock.Body()
	backendBody.SetAttributeValue("bucket", cty.StringVal(b.Bucket))
	backendBody.SetAttributeValue("key", cty.StringVal(fmt.Sprintf("%s/%s", module.InfraName(), module.Name())))
	backendBody.SetAttributeValue("region", cty.StringVal(b.Region))

	terraformBlock.Body().SetAttributeValue("required_version", cty.StringVal("~> 0.13"))
	return f.Bytes(), nil

}

// GetRemoteStateHCL generate terraform remote state for this backend.
func (b *BackendDo) GetRemoteStateHCL(module project.Module) ([]byte, error) {
	f := hclwrite.NewEmptyFile()

	rootBody := f.Body()
	dataBlock := rootBody.AppendNewBlock("data", []string{"terraform_remote_state", fmt.Sprintf("%s-%s", module.InfraName(), module.Name())})
	dataBody := dataBlock.Body()
	dataBody.SetAttributeValue("backend", cty.StringVal("s3"))
	dataBody.SetAttributeValue("config", cty.MapVal(map[string]cty.Value{
		"bucket": cty.StringVal(b.Bucket),
		"key":    cty.StringVal(fmt.Sprintf("%s/%s", module.InfraName(), module.Name())),
		"region": cty.StringVal(b.Region),
	}))

	return f.Bytes(), nil
}
