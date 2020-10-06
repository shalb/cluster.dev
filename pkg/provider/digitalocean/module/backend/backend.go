package backend

import (
	"fmt"
	"path/filepath"

	"github.com/apex/log"
	"github.com/shalb/cluster.dev/internal/config"
	"github.com/shalb/cluster.dev/internal/executor"
	"github.com/shalb/cluster.dev/pkg/cluster"
	"github.com/shalb/cluster.dev/pkg/provider"
	"github.com/shalb/cluster.dev/pkg/provider/digitalocean"
	"github.com/shalb/cluster.dev/pkg/provider/digitalocean/provisioner"
)

const moduleName = "backend"

// Type for module tfVars JSON.
type backendVarsSpec struct {
	Region string `json:"region"`
	Bucket string `json:"do_spaces_backend_bucket"`
}

// Backend type for s3 backend module.
type Backend struct {
	config    backendVarsSpec
	terraform *executor.TerraformRunner
	bash      *executor.BashRunner
	moduleDir string
}

func init() {
	err := digitalocean.RegisterActivityFactory("modules", moduleName, &Factory{})
	if err != nil {
		log.Fatalf("can't register digitalocean backend module")
	}
}

// Factory create new backend module.
type Factory struct{}

// New create new backend instance.
func (f *Factory) New(providerConf digitalocean.Config, clusterState *cluster.State) (provider.Activity, error) {

	log.Debugf("Init backend module wit provider config: %+v", providerConf)
	backend := &Backend{}
	backend.moduleDir = filepath.Join(config.Global.ProjectRoot, "terraform/digitalocean/"+moduleName)
	backend.config.Bucket = providerConf.ClusterName
	backend.config.Region = providerConf.Region
	var err error
	backend.terraform, err = executor.NewTerraformRunner(backend.moduleDir, provisioner.GetAwsAuthEnv()...)
	if err != nil {
		return nil, err
	}
	// Create bash runner and add AWS env for s3cmd auth. See GetAwsAuthEnv().
	backend.bash, err = executor.NewBashRunner(backend.moduleDir, provisioner.GetAwsAuthEnv()...)
	if err != nil {
		return nil, err
	}
	backend.terraform.LogLabels = append(backend.terraform.LogLabels, fmt.Sprintf("cluster='%s'", providerConf.ClusterName))
	backend.bash.LogLabels = append(backend.bash.LogLabels, fmt.Sprintf("cluster='%s'", providerConf.ClusterName))
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
	if err := s.terraform.Clear(); err != nil {
		return err
	}
	// Init terraform without backend spec (empty spec).
	if err := s.terraform.Init(digitalocean.BackendSpec{}); err != nil {
		return err
	}
	// Apply. Create backend.
	if err := s.terraform.Apply(s.config); err != nil {
		return err
	}
	return nil
}

// Destroy - remove s3 bucket.
func (s *Backend) Destroy() error {
	log.Debug("Delete s3 bucket.")
	// Set variables.
	command := fmt.Sprintf("s3cmd rb \"s3://%s\" --host='%s.digitaloceanspaces.com' --host-bucket='%%(bucket)s.%s.digitaloceanspaces.com' --recursive --force", s.config.Bucket, s.config.Region, s.config.Region)
	err := s.bash.Run(command)
	if err != nil {
		log.Warnf("Can't mark bucket objects: %s", err.Error())
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
	err = s.terraform.Init(digitalocean.BackendSpec{})
	err = s.terraform.Import(s.config, "digitalocean_spaces_bucket.terraform_state", s.config.Region+","+s.config.Bucket)
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
