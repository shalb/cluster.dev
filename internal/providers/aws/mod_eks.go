package aws

import (
	"fmt"

	"github.com/apex/log"
	"github.com/romanprog/c-dev/executor"
)

// Variables set for eks module tfvars.
type eksVarsSpec struct {
	AvailabilityZones          []string    `json:"availability_zones"`
	Region                     string      `json:"region"`
	ClusterName                string      `json:"cluster_name"`
	VpcID                      string      `json:"vpc_id"`
	ClusterVersion             string      `json:"cluster_version" yaml:"version"`
	WorkerAdditionalSgIds      []string    `json:"worker_additional_security_group_ids,omitempty"`
	WorkerSubnetType           string      `json:"workers_subnets_type"`
	WorkerGroupsLaunchTemplate interface{} `json:"worker_groups_launch_template,omitempty"`
}

// Eks - type for eks module.
type Eks struct {
	config      eksVarsSpec
	backendConf executor.BackendSpec
	terraform   *executor.TerraformRunner
	backendKey  string
	moduleDir   string
}

// NewEks create new eks module instance (go instance).
func NewEks(providerConf providerConfSpec) (*Eks, error) {
	var eks Eks
	eks.moduleDir = "terraform/aws/eks"
	eks.backendKey = "states/terraform-k8s.state"
	eks.backendConf = executor.BackendSpec{
		Bucket: providerConf.ClusterName,
		Key:    eks.backendKey,
		Region: providerConf.Region,
	}
	// TODO: Remove this hack.
	// Hack: add "public_ip" = "true" to all node groups when use the default VPC.
	if nodeGroups, ok := providerConf.Provisioner["node_group"]; ok {
		switch nodeGroups.(type) {
		case []interface{}:
			if providerConf.Vpc == "default" {
				for index := range nodeGroups.([]interface{}) {
					nodeGroups.([]interface{})[index].(map[interface{}]interface{})["public_ip"] = "true"
				}
			}
		default:
			return nil, fmt.Errorf("'node_group' field must be an array type")
		}
	}

	version, _ := providerConf.Provisioner["version"]
	eks.config.AvailabilityZones = providerConf.AvailabilityZones
	eks.config.ClusterName = providerConf.ClusterName
	eks.config.ClusterVersion = fmt.Sprintf("%v", version)
	eks.config.Region = providerConf.Region
	eks.config.VpcID = providerConf.Vpc
	eks.config.WorkerGroupsLaunchTemplate = providerConf.Provisioner["node_group"]

	var err error
	eks.terraform, err = executor.NewTerraformRunner(eks.moduleDir)
	if err != nil {
		return nil, err
	}
	return &eks, nil
}

// Deploy eks.
func (s *Eks) Deploy() error {
	// Clear terraform cache tmp dir.
	log.Debug("Terraform init/plan.")
	err := s.terraform.Clear()
	if err != nil {
		return err
	}
	// Init terraform with backend config.
	err = s.terraform.Init(s.backendConf)
	if err != nil {
		return err
	}
	// Plan.
	err = s.terraform.Plan(s.config, "-compact-warnings", "-out=tfplan")
	if err != nil {
		return err
	}
	// Apply plan.
	err = s.terraform.ApplyPlan("tfplan", "-compact-warnings")
	if err != nil {
		return err
	}
	return nil
}

// Destroy - remove eks.
func (s *Eks) Destroy() error {
	// Init terraform.
	err := s.terraform.Init(s.backendConf)
	if err != nil {
		return err
	}
	// Destroy.
	return s.terraform.Destroy(s.config, "-compact-warnings")
}

// Check - do nothing, not used yet.
func (s *Eks) Check() (bool, error) {
	return true, nil
}

// ModulePath - return terraform module path.
func (s *Eks) ModulePath() string {
	return s.moduleDir
}
