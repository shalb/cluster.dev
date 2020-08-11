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
	s3 := &Backend{}
	s3.moduleDir = filepath.Join(config.Global.ProjectRoot, "terraform/digitalocean/"+moduleName)
	s3.config.Bucket = providerConf.ClusterName
	s3.config.Region = providerConf.Region
	var err error
	s3.terraform, err = executor.NewTerraformRunner(s3.moduleDir)
	if err != nil {
		return nil, err
	}
	// Create bash runner and add AWS env for s3cmd auth. See GetAwsAuthEnv().
	s3.bash, err = executor.NewBashRunner(s3.moduleDir, provisioner.GetAwsAuthEnv()...)
	if err != nil {
		return nil, err
	}
	return s3, nil
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
	command := fmt.Sprintf("s3cmd rb \"s3://%s\" --host='$cluster_cloud_region.digitaloceanspaces.com' --host-bucket='%%(bucket)s.$cluster_cloud_region.digitaloceanspaces.com --recursive --force", s.config.Bucket)
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
