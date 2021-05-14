package s3

import (
	"fmt"
	"time"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/shalb/cluster.dev/pkg/aws"
	"github.com/shalb/cluster.dev/pkg/project"
	"github.com/shalb/cluster.dev/pkg/utils"
	"github.com/zclconf/go-cty/cty"
)

// Backend - describe s3 backend for interface package.backend.
type Backend struct {
	name       string
	Bucket     string `yaml:"bucket"`
	Region     string `yaml:"region"`
	ProjectPtr *project.Project
	state      map[string]interface{}
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
	return "s3"
}

// type backendConfigSpec struct {
// 	Bucket string `hcl:"bucket"`
// 	Key    string `hcl:"key"`
// 	Region string `hcl:"region"`
// }

// GetBackendBytes generate terraform backend config.
func (b *Backend) GetBackendBytes(infraName, moduleName string) ([]byte, error) {
	f, err := b.GetBackendHCL(infraName, moduleName)
	if err != nil {
		return nil, err
	}
	return f.Bytes(), nil
}

// GetBackendHCL generate terraform backend config.
func (b *Backend) GetBackendHCL(infraName, moduleName string) (*hclwrite.File, error) {
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
func (b *Backend) GetRemoteStateHCL(infraName, moduleName string) ([]byte, error) {
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

func (b *Backend) LockState() error {
	// Check if state file exists.
	lockKey := fmt.Sprintf("cdev.%s.lock", b.ProjectPtr.Name())
	_, err := aws.S3Get(b.Region, b.Bucket, lockKey)
	if err == nil {
		return fmt.Errorf("lock state file found, the state is locked")
	}

	sessionID := utils.RandString(10)
	err = aws.S3Put(b.Region, b.Bucket, lockKey, sessionID)
	if err != nil {
		return fmt.Errorf("can't save lock state file: %v", err.Error())
	}
	time.Sleep(time.Millisecond * 500)
	id, err := aws.S3Get(b.Region, b.Bucket, lockKey)
	if err != nil {
		return fmt.Errorf("can't save lock state file. Internal error")
	}
	if id != sessionID {
		return fmt.Errorf("state is locked by another process")
	}
	return nil
}

func (b *Backend) UnlockState() error {
	lockKey := fmt.Sprintf("cdev.%s.lock", b.ProjectPtr.Name())
	return aws.S3Delete(b.Region, b.Bucket, lockKey)
}

func (b *Backend) WriteState(stateData string) error {
	stateKey := fmt.Sprintf("cdev.%s.state", b.ProjectPtr.Name())
	err := aws.S3Put(b.Region, b.Bucket, stateKey, stateData)
	if err != nil {
		return fmt.Errorf("can't save state file: %v", err.Error())
	}
	return nil
}

func (b *Backend) ReadState() (string, error) {
	stateKey := fmt.Sprintf("cdev.%s.state", b.ProjectPtr.Name())
	return aws.S3Get(b.Region, b.Bucket, stateKey)
}
