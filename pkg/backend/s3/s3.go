package s3

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/apex/log"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/ec2/imds"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	awsbase "github.com/hashicorp/aws-sdk-go-base/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/shalb/cluster.dev/pkg/config"
	"github.com/shalb/cluster.dev/pkg/project"
	"github.com/shalb/cluster.dev/pkg/utils"
	"github.com/zclconf/go-cty/cty"
	"gopkg.in/yaml.v3"
)

// Backend - describe s3 backend for interface package.backend.
type Backend struct {
	name     string     `yaml:"-"`
	s3Client *s3.Client `yaml:"-"`

	Bucket                    string `yaml:"bucket"`
	Region                    string `yaml:"region"`
	AccessKey                 string `yaml:"access_key,omitempty"`
	SecretKey                 string `yaml:"secret_key,omitempty"`
	Profile                   string `yaml:"profile,omitempty"`
	Token                     string `yaml:"token,omitempty"`
	Endpoint                  string `yaml:"endpoint,omitempty"`
	SkipMetadataApiCheck      bool   `yaml:"skip_metadata_api_check,omitempty"`
	SkipCredentialsValidation bool   `yaml:"skip_credentials_validation,omitempty"`
	MaxRetries                int    `yaml:"max_retries,omitempty"`
	SharedCredentialsFile     string `yaml:"shared_credentials_file,omitempty"`
	SkipRegionValidation      string `yaml:"skip_region_validation,omitempty"`
	StsEndpoint               string `yaml:"sts_endpoint,omitempty"`
	IamEndpoint               string `yaml:"iam_endpoint,omitempty"`
	ForcePathStyle            bool   `yaml:"force_path_style,omitempty"`
	// Assume role
	AssumeRoleDurationSeconds   string            `yaml:"assume_role_duration_seconds,omitempty"`
	AssumeRolePolicy            string            `yaml:"assume_role_policy,omitempty"`
	AssumeRolePolicyArns        []string          `yaml:"assume_role_policy_arns,omitempty"`
	AssumeRoleTags              map[string]string `yaml:"assume_role_tags,omitempty"`
	AssumeRoleTransitiveTagKeys string            `yaml:"assume_role_transitive_tag_keys,omitempty"`
	ExternalId                  string            `yaml:"external_id,omitempty"`
	RoleArn                     string            `yaml:"role_arn,omitempty"`
	SessionName                 string            `yaml:"session_name,omitempty"`

	Workspaces string `yaml:"workspaces,omitempty"`

	ProjectPtr *project.Project       `yaml:"-"`
	state      map[string]interface{} `yaml:"-"`
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

// GetBackendBytes generate terraform backend config.
func (b *Backend) GetBackendBytes(stackName, unitName string) ([]byte, error) {
	f, err := b.GetBackendHCL(stackName, unitName)
	if err != nil {
		return nil, err
	}
	return f.Bytes(), nil
}

func (b *Backend) Configure() error {
	var ar *awsbase.AssumeRole
	// Configure assume role
	if b.RoleArn != "" || b.ExternalId != "" {
		ar = &awsbase.AssumeRole{
			RoleARN:     b.RoleArn,
			ExternalID:  b.ExternalId,
			Policy:      b.AssumeRolePolicy,
			PolicyARNs:  b.AssumeRolePolicyArns,
			Tags:        b.AssumeRoleTags,
			SessionName: b.SessionName,
		}
		if b.AssumeRoleDurationSeconds != "" {
			duration, err := time.ParseDuration(b.AssumeRoleDurationSeconds)
			if err != nil {
				return fmt.Errorf("configure s3 backend: %w", err)
			}
			ar.Duration = duration
		}
	}
	var sharedCredentialsFiles []string
	if b.SharedCredentialsFile != "" {
		sharedCredentialsFiles = []string{
			b.SharedCredentialsFile,
		}
	}
	cfg := &awsbase.Config{
		AccessKey:              b.AccessKey,
		SecretKey:              b.SecretKey,
		Profile:                b.Profile,
		Region:                 b.Region,
		SkipCredsValidation:    b.SkipCredentialsValidation,
		MaxRetries:             b.MaxRetries,
		AssumeRole:             ar,
		Token:                  b.Token,
		IamEndpoint:            b.IamEndpoint,
		SharedCredentialsFiles: sharedCredentialsFiles,
		APNInfo:                stdUserAgentProducts(),
	}
	if b.SkipMetadataApiCheck {
		cfg.EC2MetadataServiceEnableState = imds.ClientDisabled
	} else {
		cfg.EC2MetadataServiceEnableState = imds.ClientEnabled
	}

	ctx := context.TODO()
	_, awsConfig, diags := awsbase.GetAwsConfig(ctx, cfg)

	if diags.HasError() {
		diagError := fmt.Sprintln("s3 configuration diagnostics returns errors:")
		for _, diag := range diags {
			diagError = fmt.Sprintf("%s\nSummary: %s\nDetails: %s", diagError, diag.Summary(), diag.Detail())
		}
		return fmt.Errorf(diagError)
	}

	b.s3Client = s3.NewFromConfig(awsConfig,
		func(o *s3.Options) {
			if b.ForcePathStyle {
				o.UsePathStyle = true
			}
			if b.Endpoint != "" {
				o.BaseEndpoint = aws.String(b.Endpoint)
			}
		})
	return nil
}

func stdUserAgentProducts() *awsbase.APNInfo {
	return &awsbase.APNInfo{
		PartnerName: "Shalb",
		Products: []awsbase.UserAgentProduct{
			{Name: "Cluster.dev", Version: config.Global.Version, Comment: "+https://cluster.dev"},
		},
	}
}

// GetBackendHCL generate terraform backend config.
func (b *Backend) GetBackendHCL(stackName, unitName string) (*hclwrite.File, error) {
	f := hclwrite.NewEmptyFile()
	rootBody := f.Body()
	terraformBlock := rootBody.AppendNewBlock("terraform", []string{})
	backendBlock := terraformBlock.Body().AppendNewBlock("backend", []string{"s3"})
	backendBody := backendBlock.Body()
	// backendBody.SetAttributeValue("bucket", cty.StringVal(b.Bucket))
	backendBody.SetAttributeValue("key", cty.StringVal(fmt.Sprintf("%s/%s.state", stackName, unitName)))
	// backendBody.SetAttributeValue("region", cty.StringVal(b.Region))
	terraformBlock.Body().SetAttributeValue("required_version", cty.StringVal("~> 0.13"))
	bkMap, err := getBackendMap(*b)
	if err != nil {
		return nil, err
	}
	for key, val := range bkMap {
		backendBody.SetAttributeValue(key, val)
	}
	return f, nil

}

func getBackendMap(in Backend) (res map[string]cty.Value, err error) {
	tmpData, err := yaml.Marshal(in)
	if err != nil {
		return
	}
	resMap := map[string]interface{}{}
	err = yaml.Unmarshal(tmpData, &resMap)
	err = utils.ResolveYamlError(tmpData, err)
	res = map[string]cty.Value{}
	for k, v := range resMap {
		res[k] = cty.StringVal(fmt.Sprintf("%v", v))
	}
	return
}

// GetRemoteStateHCL generate terraform remote state for this backend.
func (b *Backend) GetRemoteStateHCL(stackName, unitName string) ([]byte, error) {
	f := hclwrite.NewEmptyFile()

	rootBody := f.Body()
	dataBlock := rootBody.AppendNewBlock("data", []string{"terraform_remote_state", fmt.Sprintf("%s-%s", stackName, unitName)})
	dataBody := dataBlock.Body()
	dataBody.SetAttributeValue("backend", cty.StringVal("s3"))
	config, err := getBackendMap(*b)
	if err != nil {
		return nil, fmt.Errorf("generate s3 remote state: %w", err)
	}
	config["key"] = cty.StringVal(fmt.Sprintf("%s/%s.state", stackName, unitName))

	dataBody.SetAttributeValue("config", cty.MapVal(config))
	return f.Bytes(), nil
}

func (b *Backend) ReadState() (string, error) {
	result, err := b.s3Client.GetObject(
		context.TODO(),
		&s3.GetObjectInput{
			Bucket: &b.Bucket,
			Key:    b.stateKey(),
		},
	)
	if err != nil {
		var bne *types.NoSuchKey
		if errors.As(err, &bne) {
			return "", nil
		}
		return "", fmt.Errorf("get state from s3 bucket: %v", err.Error())
	}
	defer result.Body.Close()
	body, err := io.ReadAll(result.Body)
	if err != nil {
		return "", fmt.Errorf("get state from s3 bucket: read file body: %v", err.Error())
	}
	return string(body), nil
}

func (b *Backend) WriteState(stateData string) error {
	ioBody := strings.NewReader(stateData)
	_, err := b.s3Client.PutObject(
		context.TODO(),
		&s3.PutObjectInput{
			Bucket: &b.Bucket,
			Key:    b.stateKey(),
			Body:   ioBody,
		},
	)
	if err != nil {
		return fmt.Errorf("write state to s3 bucket: %v", err.Error())
	}
	return nil
}

func (b *Backend) stateKey() *string {
	res := fmt.Sprintf("cdev.%s.state", b.ProjectPtr.Name())
	return &res
}

func (b *Backend) lockKey() *string {
	res := fmt.Sprintf("cdev.%s.lock", b.ProjectPtr.Name())
	return &res
}

func (b *Backend) LockState() error {
	// Check if state file exists.
	log.Debugf("Locking s3 state. Project: '%v', bucket: '%v'", b.ProjectPtr.Name(), b.Bucket)
	result, err := b.s3Client.GetObject(
		context.TODO(),
		&s3.GetObjectInput{
			Bucket: &b.Bucket,
			Key:    b.lockKey(),
		},
	)
	if err != nil {
		var bne *types.NoSuchKey
		if !errors.As(err, &bne) {
			return fmt.Errorf("lock state: get lock file: %v", err.Error())
		}
	} else {
		defer result.Body.Close()
	}
	if err == nil {
		return fmt.Errorf("lock state file found, the state is locked. Use command 'cdev state unlock' to force unlock (unsafe)")
	}
	// Lock file not found, process lock
	ioBody := strings.NewReader(b.ProjectPtr.SessionId)

	_, err = b.s3Client.PutObject(
		context.TODO(),
		&s3.PutObjectInput{
			Bucket: &b.Bucket,
			Key:    b.lockKey(),
			Body:   ioBody,
		},
	)
	if err != nil {
		return fmt.Errorf("write state to s3 bucket: %v", err.Error())
	}
	return nil
}

func (b *Backend) UnlockState() error {
	log.Debugf("Unlocking s3 state. Project: '%v', bucket: '%v'", b.ProjectPtr.Name(), b.Bucket)
	_, err := b.s3Client.DeleteObject(
		context.TODO(),
		&s3.DeleteObjectInput{
			Bucket: &b.Bucket,
			Key:    b.lockKey(),
		},
	)
	if err != nil {
		return fmt.Errorf("delete state lock from s3 bucket: %v", err.Error())
	}
	return nil
}
