package gcs

import (
	"cloud.google.com/go/storage"
	"context"
	"fmt"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/shalb/cluster.dev/pkg/hcltools"
	"github.com/shalb/cluster.dev/pkg/project"
	"github.com/shalb/cluster.dev/pkg/utils"
	"github.com/zclconf/go-cty/cty"
	"gopkg.in/yaml.v3"
	"io/ioutil"
)

// Backend - describe GCS backend for interface package.backend.
type Backend struct {
	name        string                 `yaml:"-"`
	Bucket      string                 `yaml:"bucket"`
	Credentials string                 `yaml:"credentials,omitempty"`
	Prefix      string                 `yaml:"prefix"`
	state       map[string]interface{} `yaml:"-"`
	ProjectPtr  *project.Project       `yaml:"-"`
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
	err = utils.ResolveYamlError(tmpData, err)
	return
}

func (b *Backend) LockState() error {
	lockKey := fmt.Sprintf("cdev.%s.lock", b.ProjectPtr.Name())

	// Create a GCS client.
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return err
	}

	// Check if the lock object exists.
	_, err = client.Bucket(b.Bucket).Object(lockKey).Attrs(ctx)
	if err == nil {
		return fmt.Errorf("lock state file found, the state is locked")
	}

	sessionID := utils.RandString(10)

	// Create the lock object with the sessionID.
	lockObject := client.Bucket(b.Bucket).Object(lockKey)
	w := lockObject.NewWriter(ctx)
	defer w.Close()

	if _, err := w.Write([]byte(sessionID)); err != nil {
		return fmt.Errorf("can't save lock state file: %v", err.Error())
	}

	// Sleep and read the sessionID from the lock object.
	// Compare it with the generated sessionID.

	return nil
}
func (b *Backend) UnlockState() error {
	lockKey := fmt.Sprintf("cdev.%s.lock", b.ProjectPtr.Name())

	// Create a GCS client.
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return err
	}

	// Delete the lock object.
	return client.Bucket(b.Bucket).Object(lockKey).Delete(ctx)
}

func (b *Backend) WriteState(stateData string) error {
	stateKey := fmt.Sprintf("cdev.%s.state", b.ProjectPtr.Name())

	// Create a GCS client.
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return err
	}

	// Create or overwrite the state object with stateData.
	stateObject := client.Bucket(b.Bucket).Object(stateKey)
	w := stateObject.NewWriter(ctx)
	defer w.Close()

	if _, err := w.Write([]byte(stateData)); err != nil {
		return fmt.Errorf("can't save state file: %v", err.Error())
	}

	return nil
}
func (b *Backend) ReadState() (string, error) {
	stateKey := fmt.Sprintf("cdev.%s.state", b.ProjectPtr.Name())

	// Create a GCS client.
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return "", err
	}

	// Check if the object exists.
	_, err = client.Bucket(b.Bucket).Object(stateKey).Attrs(ctx)
	if err != nil {
		if err == storage.ErrObjectNotExist {
			// fmt.Printf("Object '%s' does not exist in bucket '%s'\n", stateKey, b.Bucket)
			return "", nil
		}
		fmt.Printf("Error checking object existence: %v\n", err)
		return "", err
	}

	// Read the state object.
	stateObject := client.Bucket(b.Bucket).Object(stateKey)
	r, err := stateObject.NewReader(ctx)
	if err != nil {
		return "", err
	}
	defer r.Close()

	stateData, err := ioutil.ReadAll(r)
	if err != nil {
		return "", err
	}

	return string(stateData), nil
}
