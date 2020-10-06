package minikube

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

// Variables set for minikube module tfvars.
type tfVars struct {
	HostedZone      string `json:"hosted_zone"`
	Region          string `json:"region"`
	ClusterName     string `json:"cluster_name"`
	AwsInstanceType string `json:"aws_instance_type"`
}

// Minikube type for minikube module instance.
type Minikube struct {
	config      tfVars
	backendConf aws.BackendSpec
	terraform   *executor.TerraformRunner
	backendKey  string
	moduleDir   string
}

func init() {
	err := aws.RegisterActivityFactory("modules", "minikube", &Factory{})
	if err != nil {
		log.Fatalf("can't register aws minikube module")
	}
}

// Factory create new minikube module.
type Factory struct{}

// New create new eks instance.
func (f *Factory) New(providerConf aws.Config, clusterState *cluster.State) (provider.Activity, error) {
	miniKube := &Minikube{}
	miniKube.moduleDir = filepath.Join(config.Global.ProjectRoot, "terraform/aws/minikube")
	miniKube.backendKey = "states/terraform-k8s.state"
	miniKube.backendConf = aws.BackendSpec{
		Bucket: providerConf.ClusterName,
		Key:    miniKube.backendKey,
		Region: providerConf.Region,
	}
	instanceType, ok := providerConf.Provisioner["instanceType"].(string)
	if !ok {
		return nil, fmt.Errorf("can't determinate instance type for minikube")
	}
	miniKube.config = tfVars{
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
	miniKube.terraform.LogLabels = append(miniKube.terraform.LogLabels, fmt.Sprintf("cluster='%s'", providerConf.ClusterName))
	return miniKube, nil
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

// Path - return module path.
func (s *Minikube) Path() string {
	return s.moduleDir
}

// Clear - remove tmp and cache files.
func (s *Minikube) Clear() error {
	return s.terraform.Clear()
}
