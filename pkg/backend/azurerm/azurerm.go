package azurerm

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/shalb/cluster.dev/pkg/hcltools"
	"github.com/shalb/cluster.dev/pkg/project"
	"github.com/shalb/cluster.dev/pkg/utils"
	"github.com/zclconf/go-cty/cty"
	"gopkg.in/yaml.v3"
)

// Backend - describe azure backend for interface package.backend.
type Backend struct {
	client                        *azblob.Client
	name                          string                 `yaml:"-"`
	state                         map[string]interface{} `yaml:"-"`
	ProjectPtr                    *project.Project       `yaml:"-"`
	ContainerName                 string                 `yaml:"container_name,omitempty"`
	StorageAccountName            string                 `yaml:"storage_account_name,omitempty"`
	ResourceGroupName             string                 `yaml:"resource_group_name,omitempty"`
	AccessKey                     string                 `yaml:"access_key,omitempty"`
	ClientID                      string                 `yaml:"client_id,omitempty"`
	ClientCertificatePassword     string                 `yaml:"client_certificate_password,omitempty"`
	ClientCertificatePath         string                 `yaml:"client_certificate_path,omitempty"`
	ClientSecret                  string                 `yaml:"client_secret,omitempty"`
	CustomResourceManagerEndpoint string                 `yaml:"endpoint,omitempty"`
	MetadataHost                  string                 `yaml:"metadata_host,omitempty"`
	Environment                   string                 `yaml:"environment,omitempty"`
	MsiEndpoint                   string                 `yaml:"msi_endpoint,omitempty"`
	OIDCToken                     string                 `yaml:"oidc_token,omitempty"`
	OIDCTokenFilePath             string                 `yaml:"oidc_token_file_path,omitempty"`
	OIDCRequestURL                string                 `yaml:"oidc_request_url,omitempty"`
	OIDCRequestToken              string                 `yaml:"oidc_request_token,omitempty"`
	SasToken                      string                 `yaml:"sas_token,omitempty"`
	SubscriptionID                string                 `yaml:"subscription_id,omitempty"`
	TenantID                      string                 `yaml:"tenant_id,omitempty"`
	UseMsi                        bool                   `yaml:"use_msi,omitempty"`
	UseOIDC                       bool                   `yaml:"use_oidc,omitempty"`
	UseAzureADAuthentication      bool                   `yaml:"use_azuread_auth,omitempty"`
}

func (b *Backend) Configure() error {
	if b.client != nil {
		return nil
	}

	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return err
	}

	client, err := azblob.NewClient("https://"+b.StorageAccountName+".blob.core.windows.net/", cred, nil)
	if err != nil {
		return err
	}

	b.client = client
	return nil
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
	return "azurerm"
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
	bConfigTmpl["key"] = fmt.Sprintf("%s-%s.state", stackName, unitName)
	f := hclwrite.NewEmptyFile()
	rootBody := f.Body()
	terraformBlock := rootBody.AppendNewBlock("terraform", []string{})
	backendBlock := terraformBlock.Body().AppendNewBlock("backend", []string{"azurerm"})
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
	bConfigTmpl["key"] = fmt.Sprintf("%s-%s.state", stackName, unitName)
	f := hclwrite.NewEmptyFile()

	rootBody := f.Body()
	dataBlock := rootBody.AppendNewBlock("data", []string{"terraform_remote_state", fmt.Sprintf("%s-%s", stackName, unitName)})
	dataBody := dataBlock.Body()
	dataBody.SetAttributeValue("backend", cty.StringVal("azurerm"))
	config, err := hcltools.InterfaceToCty(bConfigTmpl)
	if err != nil {
		return nil, err
	}
	dataBody.SetAttributeValue("config", config)
	return f.Bytes(), nil
}

func (b *Backend) LockState() error {
	lockKey := fmt.Sprintf("cdev.%s.lock", b.ProjectPtr.Name())
	ctx := context.Background()

	// Check if the blob exists.
	_, err := b.client.DownloadStream(ctx, b.ContainerName, lockKey, nil)
	if err == nil {
		return fmt.Errorf("Lock state blob found, the state is locked")
	}

	sessionID := utils.RandString(10)

	// Upload a blob to Azure Blob Storage.
	buf := []byte(sessionID)
	_, err = b.client.UploadBuffer(ctx, b.ContainerName, lockKey, buf, &azblob.UploadBufferOptions{})
	if err != nil {
		return fmt.Errorf("Can't save lock state blob: %v", err)
	}

	return nil
}

func (b *Backend) UnlockState() error {
	lockKey := fmt.Sprintf("cdev.%s.lock", b.ProjectPtr.Name())
	ctx := context.Background()
	_, err := b.client.DeleteBlob(ctx, b.ContainerName, lockKey, nil)
	if err != nil {
		return fmt.Errorf("Can't unlock state: %v", err)
	}

	return nil
}

func (b *Backend) WriteState(stateData string) error {
	stateKey := fmt.Sprintf("cdev.%s.state", b.ProjectPtr.Name())
	ctx := context.Background()
	buf := []byte(stateData)
	_, err := b.client.UploadBuffer(ctx, b.ContainerName, stateKey, buf, &azblob.UploadBufferOptions{})
	if err != nil {
		return fmt.Errorf("Can't save state blob: %v", err)
	}

	return nil
}

func (b *Backend) ReadState() (string, error) {
	stateKey := fmt.Sprintf("cdev.%s.state", b.ProjectPtr.Name())
	ctx := context.Background()

	// Check if the object exists.
	_, err := b.client.DownloadStream(ctx, b.ContainerName, stateKey, nil)
	if err != nil {
		// Check if the error message contains "BlobNotFound" to identify the error.
		if strings.Contains(err.Error(), "BlobNotFound") {
			fmt.Println("The blob does not exist.")
			return "", nil
		}
		return "", err
	}

	// Download the blob
	get, err := b.client.DownloadStream(ctx, b.ContainerName, stateKey, nil)
	if err != nil {
		fmt.Errorf("Can't read state blob: %v", err)
		return "", err
	}

	stateData := bytes.Buffer{}
	retryReader := get.NewRetryReader(ctx, &azblob.RetryReaderOptions{})
	_, err = stateData.ReadFrom(retryReader)

	err = retryReader.Close()

	return stateData.String(), nil
}
