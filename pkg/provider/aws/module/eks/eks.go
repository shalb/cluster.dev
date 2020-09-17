package eks

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

// Variables set for eks module tfvars.
type tfVars struct {
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
	config      tfVars
	backendConf aws.BackendSpec
	terraform   *executor.TerraformRunner
	backendKey  string
	moduleDir   string
}

func init() {
	err := aws.RegisterActivityFactory("modules", "eks", &Factory{})
	if err != nil {
		log.Fatalf("can't register aws eks module")
	}
}

// Factory create new eks module.
type Factory struct{}

// New create new eks instance.
func (f *Factory) New(providerConf aws.Config, clusterState *cluster.State) (provider.Activity, error) {
	eks := &Eks{}
	eks.moduleDir = filepath.Join(config.Global.ProjectRoot, "terraform/aws/eks")
	eks.backendKey = "states/terraform-k8s.state"
	eks.backendConf = aws.BackendSpec{
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
	eks.terraform.LogLabels = append(eks.terraform.LogLabels, fmt.Sprintf("cluster='%s'", providerConf.ClusterName))
	return eks, nil
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

// Path - return terraform module path.
func (s *Eks) Path() string {
	return s.moduleDir
}

// Clear - remove tmp and cache files.
func (s *Eks) Clear() error {
	return s.terraform.Clear()
}
