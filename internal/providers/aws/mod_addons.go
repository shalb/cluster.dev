package aws

import (
	"fmt"

	"github.com/apex/log"
	"github.com/romanprog/c-dev/executor"
)

// Variables set for module tfvars.
type addonsVarsSpec struct {
	Region             string `json:"region"`
	ClusterName        string `json:"cluster_name"`
	ConfigPath         string `json:"config_path"`
	ClusterCloudDomain string `json:"cluster_cloud_domain"`
	Eks                string `json:"eks"`
}

// Addons - type for module instance.
type Addons struct {
	config      addonsVarsSpec
	backendConf executor.BackendSpec
	terraform   *executor.TerraformRunner
	backendKey  string
	moduleDir   string
}

// NewAddons create new addons instance.
func NewAddons(providerConf providerConfSpec) (*Addons, error) {
	var addons Addons
	// Module dir.
	addons.moduleDir = "terraform/aws/addons"
	// Module state name.
	addons.backendKey = "states/terraform-addons.state"
	// Set backend config.
	addons.backendConf = executor.BackendSpec{
		Bucket: providerConf.ClusterName,
		Key:    addons.backendKey,
		Region: providerConf.Region,
	}
	// Set tfVars.
	addons.config.ClusterName = providerConf.ClusterName
	addons.config.Region = providerConf.Region
	addons.config.ClusterCloudDomain = providerConf.Domain
	addons.config.ConfigPath = fmt.Sprintf("/tmp/kubeconfig_%s", providerConf.ClusterName)

	// Detect provisioner type for module var 'eks=(true|false)'
	provisionerType, ok := providerConf.Provisioner["type"].(string)
	if !ok {
		return nil, fmt.Errorf("can't determinate provisioner type (for 'addons' module)")
	}
	if provisionerType == "eks" {
		addons.config.Eks = "true"
	} else {
		addons.config.Eks = "false"
	}
	var err error
	// Init terraform runner in module directory.
	addons.terraform, err = executor.NewTerraformRunner(addons.moduleDir)
	if err != nil {
		return nil, err
	}
	return &addons, nil
}

// Deploy addons.
func (s *Addons) Deploy() error {
	// Clear terraform cache tmp dir.
	log.Debug("Terraform init/plan.")
	err := s.terraform.Clear()
	if err != nil {
		return err
	}
	// Init terraform with backend spec.
	err = s.terraform.Init(s.backendConf)
	if err != nil {
		return err
	}
	// Plan, output to 'tfplan' file.
	err = s.terraform.Plan(s.config, "-compact-warnings", "-out=tfplan")
	if err != nil {
		return err
	}
	// Apply 'tfplan'.
	err = s.terraform.ApplyPlan("tfplan", "-compact-warnings")
	if err != nil {
		return err
	}
	return nil
}

// Destroy - remove addons.
func (s *Addons) Destroy() error {
	// Init terraform.
	err := s.terraform.Init(s.backendConf)
	if err != nil {
		return err
	}
	// Destroy.
	return s.terraform.Destroy(s.config, "-compact-warnings")
}

// Check - do nothing, not used yet.
func (s *Addons) Check() (bool, error) {
	return true, nil
}

// ModulePath - return terraform module path.
func (s *Addons) ModulePath() string {
	return s.moduleDir
}
