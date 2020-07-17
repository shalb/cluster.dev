package aws

import (
	"fmt"

	"github.com/apex/log"
	"github.com/romanprog/c-dev/executor"
)

// Type for module tfVars JSON.
type backendVarsSpec struct {
	Region string `json:"region"`
	Bucket string `json:"s3_backend_bucket"`
}

// S3Backend type for s3 backend module.
type S3Backend struct {
	config      backendVarsSpec
	backendConf executor.BackendSpec
	terraform   *executor.TerraformRunner
	bash        *executor.BashRunner
}

const backendModulePath = "terraform/aws/backend"

// NewS3Backend create new s3 backend instance.
func NewS3Backend(providerConf providerConfSpec) (*S3Backend, error) {
	var s3 S3Backend
	s3.backendConf = executor.BackendSpec{}
	s3.config.Bucket = providerConf.ClusterName
	s3.config.Region = providerConf.Region
	var err error
	s3.terraform, err = executor.NewTerraformRunner(backendModulePath)
	if err != nil {
		return nil, err
	}
	s3.bash, err = executor.NewBashRunner(backendModulePath)
	if err != nil {
		return nil, err
	}
	return &s3, nil
}

// Deploy - create s3 bucket.
func (s *S3Backend) Deploy() error {
	// sss
	log.Debug("Terraform init/plan.")
	err := s.terraform.Clear()
	if err != nil {
		return err
	}
	// Init terraform without backend speck.
	err = s.terraform.Init(s.backendConf)
	if err != nil {
		return err
	}
	// Apply. Create backend.
	err = s.terraform.Apply(s.config)
	if err != nil {
		return err
	}
	return nil
}

// Destroy - remove s3 bucket.
func (s *S3Backend) Destroy() error {
	// sss
	log.Debug("Delete s3 bucket.")
	// Set variables.
	command := fmt.Sprintf("aws s3api delete-objects --bucket \"%[1]s\" --delete \"$(aws s3api list-object-versions --bucket %[1]s --output=json --query='{Objects: Versions[].{Key:Key,VersionId:VersionId}}')\"", s.config.Bucket)
	err := s.bash.Run(command)
	if err != nil {
		log.Warnf("Can't mark bucket objects: %s", err.Error())
	}
	command = fmt.Sprintf("aws s3api delete-objects --bucket \"%[1]s\" --delete \"$(aws s3api list-object-versions --bucket %[1]s --output=json --query='{Objects: DeleteMarkers[].{Key:Key,VersionId:VersionId}}')\"", s.config.Bucket)
	err = s.bash.Run(command)
	if err != nil {
		log.Warnf("Can't remove bucket objects: %s", err.Error())
	}
	command = fmt.Sprintf("aws s3 rb \"s3://%s\" --force", s.config.Bucket)
	err = s.bash.Run(command)
	if err != nil {
		log.Warnf("Can't remove bucket: %s", err.Error())
	}
	command = fmt.Sprintf("aws --region \"%s\" dynamodb delete-table --table-name \"%s-state\"", s.config.Region, s.config.Bucket)
	err = s.bash.Run(command)
	if err != nil {
		return fmt.Errorf("can't remove dynamodb table: %s", err.Error())
	}
	return nil
}

// Check - if s3 bucket exists.
func (s *S3Backend) Check() (bool, error) {

	log.Debug("Terraform init/plan.")
	err := s.terraform.Clear()
	if err != nil {
		return false, err
	}
	// Init terraform without backend speck.
	err = s.terraform.Init(s.backendConf)
	err = s.terraform.Import(s.config, "aws_s3_bucket.terraform_state", s.config.Bucket)
	if err != nil {
		log.Debugf("Bucket is not exists, %s", err.Error())
		return false, nil
	}
	return true, nil
}
