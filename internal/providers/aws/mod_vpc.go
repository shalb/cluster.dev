package aws

import (
	"github.com/apex/log"
	"github.com/romanprog/c-dev/executor"
)

// Variables set for vpc module tfvars.
type vpcVarsSpec struct {
	VpcID             string   `json:"vpc_id"`
	Region            string   `json:"region"`
	ClusterName       string   `json:"cluster_name"`
	VpcCIDR           string   `json:"vpc_cidr"`
	AvailabilityZones []string `json:"availability_zones"`
}

// Vpc type for vpc module instance.
type Vpc struct {
	config      vpcVarsSpec
	backendConf executor.BackendSpec
	terraform   *executor.TerraformRunner
}

const vpcModulePath = "terraform/aws/vpc"
const vpcModuleBackendKey = "states/terraform-vpc.state"

// NewVpc create new vpc instance.
func NewVpc(providerConf providerConfSpec) (*Vpc, error) {
	var bk Vpc
	bk.backendConf = executor.BackendSpec{
		Bucket: providerConf.ClusterName,
		Key:    vpcModuleBackendKey,
		Region: providerConf.Region,
	}
	bk.config = vpcVarsSpec{
		VpcID:             providerConf.Vpc,
		Region:            providerConf.Region,
		ClusterName:       providerConf.ClusterName,
		VpcCIDR:           "10.8.0.0/18",
		AvailabilityZones: providerConf.AvailabilityZones,
	}
	var err error
	bk.terraform, err = executor.NewTerraformRunner(vpcModulePath)
	if err != nil {
		return nil, err
	}
	return &bk, nil
}

// Deploy - create vpc.
func (s *Vpc) Deploy() error {
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
	// Plan.
	err = s.terraform.Plan(s.config, "-compact-warnings", "-out=tfplan")
	if err != nil {
		return err
	}
	// Apply. Create DNS.
	err = s.terraform.ApplyPlan("tfplan", "-compact-warnings")
	if err != nil {
		return err
	}
	return nil
}

// Destroy - remove vpc.
func (s *Vpc) Destroy() error {
	// Init terraform without backend speck.
	err := s.terraform.Init(s.backendConf)
	if err != nil {
		return err
	}
	// Plan.
	return s.terraform.Destroy(s.config, "-compact-warnings")

}

// Check - if s3 bucket exists.
func (s *Vpc) Check() (bool, error) {
	return true, nil
}
