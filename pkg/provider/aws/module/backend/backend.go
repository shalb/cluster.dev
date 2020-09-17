package backend

import (
	"fmt"
	"path/filepath"

	"github.com/apex/log"
	"github.com/shalb/cluster.dev/internal/config"
	"github.com/shalb/cluster.dev/internal/executor"
	"github.com/shalb/cluster.dev/pkg/cluster"
	"github.com/shalb/cluster.dev/pkg/provider"
	"github.com/shalb/cluster.dev/pkg/provider/aws"
)

// Type for module tfVars JSON.
type backendVarsSpec struct {
	Region string `json:"region"`
	Bucket string `json:"s3_backend_bucket"`
}

// Backend type for s3 backend module.
type Backend struct {
	config        backendVarsSpec
	terraform     *executor.TerraformRunner
	bash          *executor.BashRunner
	moduleDir     string
	clusterConfig aws.Config
}

func init() {
	err := aws.RegisterActivityFactory("modules", "backend", &Factory{})
	if err != nil {
		log.Fatalf("can't register aws backend module")
	}
}

// Factory create new backend module.
type Factory struct{}

// New create new eks instance.
func (f *Factory) New(providerConf aws.Config, clusterState *cluster.State) (provider.Activity, error) {

	log.Debugf("Init backend module wit provider config: %+v", providerConf)
	backend := &Backend{}
	backend.moduleDir = filepath.Join(config.Global.ProjectRoot, "terraform/aws/backend")
	backend.config.Bucket = providerConf.ClusterName
	backend.config.Region = providerConf.Region
	backend.clusterConfig = providerConf
	var err error
	backend.terraform, err = executor.NewTerraformRunner(backend.moduleDir)
	if err != nil {
		return nil, err
	}
	backend.bash, err = executor.NewBashRunner(backend.moduleDir)
	if err != nil {
		return nil, err
	}
	//s3.terraform.
	backend.bash.LogLabels = append(backend.bash.LogLabels, fmt.Sprintf("cluster='%s'", providerConf.ClusterName))
	backend.terraform.LogLabels = append(backend.terraform.LogLabels, fmt.Sprintf("cluster='%s'", providerConf.ClusterName))
	return backend, nil
}

// Deploy - create s3 bucket.
func (s *Backend) Deploy() error {
	// Check if bucket exists - no need redeploy.
	if exists, err := s.Check(); err == nil {
		if exists {
			log.Debugf("Bucket '%v' exists. Nothing to deploy.", s.config.Bucket)
			return nil
		}
		log.Debugf("Deploying s3 bucket '%v'", s.config.Bucket)
	} else {
		return err
	}

	log.Debug("Terraform init/plan.")
	err := s.terraform.Clear()
	if err != nil {
		return err
	}
	// Init terraform without backend spec (empty spec).
	err = s.terraform.Init(aws.BackendSpec{})
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
func (s *Backend) Destroy() error {
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
func (s *Backend) Check() (bool, error) {

	log.Debug("Terraform init/plan.")
	err := s.terraform.Clear()
	if err != nil {
		return false, err
	}
	// Init terraform without backend spec.

	err = s.terraform.Init(aws.BackendSpec{})
	err = s.terraform.Import(s.config, "aws_s3_bucket.terraform_state", s.config.Bucket)
	if err != nil {
		log.Debugf("Bucket is not exists, %s", err.Error())
		return false, nil
	}
	return true, nil
}

// Path - return module path.
func (s *Backend) Path() string {
	return s.moduleDir
}

// Clear - remove tmp and cache files.
func (s *Backend) Clear() error {
	return s.terraform.Clear()
}
