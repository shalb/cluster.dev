package do

import (
	"fmt"

	"github.com/zclconf/go-cty/cty"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/shalb/cluster.dev/pkg/hcltools"
	"github.com/shalb/cluster.dev/pkg/project"
)

// BackendDo - describe do spaces backend for interface package.backend.
type BackendDo struct {
	name      string
	Bucket    string `yaml:"bucket"`
	Region    string `yaml:"region"`
	AccessKey string `yaml:"access_key,omitempty"`
	SecretKey string `yaml:"secret_key,omitempty"`
}

// Name return name.
func (b *BackendDo) Name() string {
	return b.name
}

// Provider return name.
func (b *BackendDo) Provider() string {
	return "do"
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
	backendBody.SetAttributeValue("region", cty.StringVal("us-east-1"))
	backendBody.SetAttributeValue("endpoint", cty.StringVal(fmt.Sprintf("%s.digitaloceanspaces.com", b.Region)))
	backendBody.SetAttributeValue("skip_credentials_validation", cty.BoolVal(true))
	backendBody.SetAttributeValue("skip_metadata_api_check", cty.BoolVal(true))
	if b.AccessKey != "" {
		backendBody.SetAttributeValue("access_key", cty.StringVal(b.AccessKey))
		backendBody.SetAttributeValue("secret_key", cty.StringVal(b.SecretKey))
	}
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

	config := map[string]interface{}{
		"bucket":                      b.Bucket,
		"key":                         fmt.Sprintf("%s/%s", module.InfraName(), module.Name()),
		"region":                      "us-east-1",
		"endpoint":                    fmt.Sprintf("%s.digitaloceanspaces.com", b.Region),
		"skip_credentials_validation": true,
		"skip_metadata_api_check":     true,
	}
	if b.AccessKey != "" {
		config["access_key"] = b.AccessKey
		config["secret_key"] = b.SecretKey
	}
	rsBucketConf, err := hcltools.InterfaceToCty(config)
	dataBody.SetAttributeValue("config", rsBucketConf)
	if err != nil {
		return nil, err
	}
	return f.Bytes(), nil
}
