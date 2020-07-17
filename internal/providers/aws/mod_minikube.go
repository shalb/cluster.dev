package aws

import (
	"fmt"

	"github.com/apex/log"
	"github.com/romanprog/c-dev/executor"
)

// Variables set for minikube module tfvars.
type minikubeVarsSpec struct {
	HostedZone      string `json:"hosted_zone"`
	Region          string `json:"region"`
	ClusterName     string `json:"cluster_name"`
	AwsInstanceType string `json:"aws_instance_type"`
}

// Minikube type for minikube module instance.
type Minikube struct {
	config      minikubeVarsSpec
	backendConf executor.BackendSpec
	terraform   *executor.TerraformRunner
	backendKey  string
	moduleDir   string
}

// NewMinikube create new minikube instance.
func NewMinikube(providerConf providerConfSpec) (*Minikube, error) {
	var miniKube Minikube
	miniKube.moduleDir = "terraform/aws/minikube"
	miniKube.backendKey = "states/terraform-k8s.state"
	miniKube.backendConf = executor.BackendSpec{
		Bucket: providerConf.ClusterName,
		Key:    miniKube.backendKey,
		Region: providerConf.Region,
	}
	instanceType, ok := providerConf.Provisioner["instanceType"].(string)
	if !ok {
		return nil, fmt.Errorf("can't determinate instance type for minikube")
	}
	miniKube.config = minikubeVarsSpec{
		HostedZone:      fmt.Sprintf("%s.%s", providerConf.ClusterName, providerConf.Domain),
		Region:          providerConf.Region,
		ClusterName:     providerConf.ClusterName,
		AwsInstanceType: instanceType,
	}
	var err error
	miniKube.terraform, err = executor.NewTerraformRunner(miniKube.moduleDir)
	if err != nil {
		return nil, err
	}
	return &miniKube, nil
}

// Deploy - create minikube.
func (s *Minikube) Deploy() error {
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
		log.Error("ERROR fuck")
		return err
	}
	return nil
}

// Destroy - remove minikube.
func (s *Minikube) Destroy() error {
	// Init terraform without backend speck.
	err := s.terraform.Init(s.backendConf)
	if err != nil {
		return err
	}
	// Plan.
	return s.terraform.Destroy(s.config, "-compact-warnings")
}

// Check - do nothing.
func (s *Minikube) Check() (bool, error) {
	return true, nil
}

// ModulePath - return module path.
func (s *Minikube) ModulePath() string {
	return s.moduleDir
}
