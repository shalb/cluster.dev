package s3

import (
	"fmt"

	"github.com/apex/log"
	"github.com/zclconf/go-cty/cty"
	"gopkg.in/yaml.v3"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/shalb/cluster.dev/pkg/project"
)

// BackendS3 - describe s3 backend for interface package.backend.
type BackendS3 struct {
	name   string
	Bucket string `yaml:"bucket"`
	Region string `yaml:"region"`
}

// Factory factory for s3 backends.
type Factory struct{}

// New creates the new s3 backend.
func (f *Factory) New(config []byte, name string) (project.Backend, error) {
	bk := BackendS3{name: name}
	err := yaml.Unmarshal(config, &bk)
	if err != nil {
		return nil, err
	}
	return &bk, nil
}

func init() {
	log.Debug("Registering backend provider s3..")
	if err := project.RegisterBackendFactory(&Factory{}, "s3"); err != nil {
		log.Trace("Can't register backend provider s3.")
	}
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

// GetBackendHCL generate terraform backend config.
func (b *BackendS3) GetBackendHCL(module project.Module) ([]byte, error) {
	f := hclwrite.NewEmptyFile()
	rootBody := f.Body()
	terraformBlock := rootBody.AppendNewBlock("terraform", []string{})
	backendBlock := terraformBlock.Body().AppendNewBlock("backend", []string{"s3"})
	backendBody := backendBlock.Body()
	backendBody.SetAttributeValue("bucket", cty.StringVal(b.Bucket))
	backendBody.SetAttributeValue("key", cty.StringVal(fmt.Sprintf("%s/%s", module.InfraPtr.Name, module.Name)))
	backendBody.SetAttributeValue("region", cty.StringVal(b.Region))

	terraformBlock.Body().SetAttributeValue("required_version", cty.StringVal("~> 0.13"))
	return f.Bytes(), nil

}

// GetRemoteStateHCL generate terraform remote state for this backend.
func (b *BackendS3) GetRemoteStateHCL(module project.Module) ([]byte, error) {
	f := hclwrite.NewEmptyFile()

	rootBody := f.Body()
	dataBlock := rootBody.AppendNewBlock("data", []string{"terraform_remote_state", fmt.Sprintf("%s-%s", module.InfraPtr.Name, module.Name)})
	dataBody := dataBlock.Body()
	dataBody.SetAttributeValue("backend", cty.StringVal("s3"))
	dataBody.SetAttributeValue("config", cty.MapVal(map[string]cty.Value{
		"bucket": cty.StringVal(b.Bucket),
		"key":    cty.StringVal(fmt.Sprintf("%s/%s", module.InfraPtr.Name, module.Name)),
		"region": cty.StringVal(b.Region),
	}))

	return f.Bytes(), nil
}
