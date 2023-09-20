package gcs

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"cloud.google.com/go/storage"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/shalb/cluster.dev/pkg/hcltools"
	"github.com/shalb/cluster.dev/pkg/project"
	"github.com/shalb/cluster.dev/pkg/utils"
	"github.com/zclconf/go-cty/cty"
	"golang.org/x/oauth2"
	"google.golang.org/api/impersonate"
	"google.golang.org/api/option"
	"gopkg.in/yaml.v3"
)

// Backend - describe GCS backend for interface package.backend.
type Backend struct {
	storageClient          *storage.Client `yaml:"-"`
	storageContext         context.Context
	name                   string                 `yaml:"-"`
	Bucket                 string                 `yaml:"bucket"`
	Credentials            string                 `yaml:"credentials,omitempty"`
	ImpersonateSA          string                 `yaml:"impersonate_service_account,omitempty"`
	ImpersonateSADelegates []string               `yaml:"impersonate_service_account_delegates,omitempty"`
	AccessToken            string                 `yaml:"access_token,omitempty"`
	Prefix                 string                 `yaml:"prefix"`
	encryptionKey          []byte                 `yaml:"encryption_key,omitempty"`
	StorageCustomEndpoint  string                 `yaml:"storage_custom_endpoint,omitempty"`
	state                  map[string]interface{} `yaml:"-"`
	ProjectPtr             *project.Project       `yaml:"-"`
}

func (b *Backend) Configure() error {
	if b.storageClient != nil {
		return nil
	}

	ctx := context.Background()
	b.storageContext = ctx

	b.Prefix = strings.TrimLeft(b.Prefix, "/")
	if b.Prefix != "" && !strings.HasSuffix(b.Prefix, "/") {
		b.Prefix = b.Prefix + "/"
	}

	var opts []option.ClientOption
	var credOptions []option.ClientOption

	// Add credential source
	var creds string
	var tokenSource oauth2.TokenSource

	if b.AccessToken != "" {
		tokenSource = oauth2.StaticTokenSource(&oauth2.Token{
			AccessToken: b.AccessToken,
		})
	} else if b.Credentials != "" {
		creds = b.Credentials
	} else if v := os.Getenv("GOOGLE_BACKEND_CREDENTIALS"); v != "" {
		creds = v
	} else {
		creds = os.Getenv("GOOGLE_CREDENTIALS")
	}

	if tokenSource != nil {
		credOptions = append(credOptions, option.WithTokenSource(tokenSource))
	} else if creds != "" {
		contents, err := ReadPathOrContents(creds)
		if err != nil {
			return fmt.Errorf("error loading credentials: %s", err.Error())
		}

		if !json.Valid([]byte(contents)) {
			return fmt.Errorf("the string provided in credentials is neither valid json nor a valid file path")
		}

		credOptions = append(credOptions, option.WithCredentialsJSON([]byte(contents)))
	}

	if b.ImpersonateSA != "" {
		ts, err := impersonate.CredentialsTokenSource(ctx, impersonate.CredentialsConfig{
			TargetPrincipal: b.ImpersonateSA,
			Scopes:          []string{storage.ScopeReadWrite},
			Delegates:       b.ImpersonateSADelegates,
		}, credOptions...)
		if err != nil {
			return err
		}
		opts = append(opts, option.WithTokenSource(ts))
	} else {
		opts = append(opts, credOptions...)
	}

	if b.StorageCustomEndpoint != "" {
		endpoint := option.WithEndpoint(b.StorageCustomEndpoint)
		opts = append(opts, endpoint)
	}
	client, err := storage.NewClient(b.storageContext, opts...)
	if err != nil {
		return fmt.Errorf("storage.NewClient() failed: %v", err.Error())
	}

	b.storageClient = client

	key := b.encryptionKey

	if len(key) > 0 {
		kc, err := ReadPathOrContents(string(key))
		if err != nil {
			return fmt.Errorf("error loading encryption key: %s", err.Error())
		}

		k, err := base64.StdEncoding.DecodeString(kc)
		if err != nil {
			return fmt.Errorf("error decoding encryption key: %s", err.Error())
		}
		b.encryptionKey = k
	}

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

	// Create a context.
	ctx := context.Background()

	// Check if the lock object exists.
	_, err := b.storageClient.Bucket(b.Bucket).Object(lockKey).Attrs(ctx)
	if err == nil {
		return fmt.Errorf("lock state file found, the state is locked")
	}

	sessionID := utils.RandString(10)

	// Create the lock object with the sessionID.
	lockObject := b.storageClient.Bucket(b.Bucket).Object(lockKey)
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

	// Create a context.
	ctx := context.Background()

	// Delete the lock object.
	return b.storageClient.Bucket(b.Bucket).Object(lockKey).Delete(ctx)
}

func (b *Backend) WriteState(stateData string) error {
	stateKey := fmt.Sprintf("cdev.%s.state", b.ProjectPtr.Name())

	// Create a context.
	ctx := context.Background()

	// Create or overwrite the state object with stateData.
	stateObject := b.storageClient.Bucket(b.Bucket).Object(stateKey)
	w := stateObject.NewWriter(ctx)
	defer w.Close()

	if _, err := w.Write([]byte(stateData)); err != nil {
		return fmt.Errorf("can't save state file: %v", err.Error())
	}

	return nil
}
func (b *Backend) ReadState() (string, error) {
	stateKey := fmt.Sprintf("cdev.%s.state", b.ProjectPtr.Name())

	// Create a context.
	ctx := context.Background()

	// Check if the object exists.
	_, err := b.storageClient.Bucket(b.Bucket).Object(stateKey).Attrs(ctx)
	if err != nil {
		if err == storage.ErrObjectNotExist {
			// fmt.Printf("Object '%s' does not exist in bucket '%s'\n", stateKey, b.Bucket)
			return "", nil
		}
		fmt.Printf("Error checking object existence: %v\n", err)
		return "", err
	}

	// Read the state object.
	stateObject := b.storageClient.Bucket(b.Bucket).Object(stateKey)
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

// ReadPathOrContents reads the contents of a file if the input is a file path,
// or returns the input as is if it's not a file path.
func ReadPathOrContents(input string) (string, error) {
	// Check if the input contains a path separator (e.g., '/').
	if strings.Contains(input, "/") {
		// Treat input as a file path and read its contents.
		contents, err := readFileContents(input)
		if err != nil {
			return "", err
		}
		return contents, nil
	}

	// Input is not a file path, return it as is.
	return input, nil
}

// readFileContents reads the contents of a file given its path.
func readFileContents(filePath string) (string, error) {
	// Read the file contents.
	contents, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(contents), nil
}
