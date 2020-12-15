package s3

import (
	"fmt"

	"github.com/apex/log"
	"github.com/rodaine/hclencoder"
	"gopkg.in/yaml.v3"

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
	type BackendConfig struct {
		BlockKey          string `hcl:",key"`
		backendConfigSpec `hcl:",squash"`
	}

	type Terraform struct {
		Backend BackendConfig `hcl:"backend"`
		ReqVer  string        `hcl:"required_version"`
	}

	type Config struct {
		TfBlock Terraform `hcl:"terraform"`
	}

	bSpeck := backendConfigSpec{
		Bucket: b.Bucket,
		Key:    fmt.Sprintf("%s/%s", module.InfraPtr.Name, module.Name),
		Region: "us-east1",
	}

	tf := Terraform{
		Backend: BackendConfig{
			BlockKey:          "s3",
			backendConfigSpec: bSpeck,
		},
		ReqVer: "~> 0.13",
	}

	input := Config{
		TfBlock: tf,
	}
	return hclencoder.Encode(input)
}

// GetRemoteStateHCL generate terraform remote state for this backend.
func (b *BackendS3) GetRemoteStateHCL(module project.Module) ([]byte, error) {

	type Data struct {
		KeyRemState  string            `hcl:",key"`
		KeyStateName string            `hcl:",key"`
		Backend      string            `hcl:"backend"`
		Config       backendConfigSpec `hcl:"config"`
	}

	type Config struct {
		TfBlock []Data `hcl:"data"`
	}

	input := Config{}
	tf := Data{
		KeyRemState:  "terraform_remote_state",
		KeyStateName: fmt.Sprintf("%s-%s", module.InfraPtr.Name, module.Name),
		Config: backendConfigSpec{
			Bucket: b.Bucket,
			Key:    fmt.Sprintf("%s/%s", module.InfraPtr.Name, module.Name),
			Region: "us-east1",
		},
		Backend: "s3",
	}

	input.TfBlock = append(input.TfBlock, tf)

	return hclencoder.Encode(input)

}
