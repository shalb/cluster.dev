package vpc

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

const myName = "vpc"

// Variables set for vpc module tfvars.
type tfVars struct {
	VpcID       string `json:"vpc_id"`
	Region      string `json:"region"`
	ClusterName string `json:"cluster_name"`
	VpcCIDR     string `json:"ip_range"`
}

// Vpc type for vpc module instance.
type Vpc struct {
	config      tfVars
	backendConf digitalocean.BackendSpec
	terraform   *executor.TerraformRunner
	backendKey  string
	moduleDir   string
}

func init() {
	err := digitalocean.RegisterActivityFactory("modules", myName, &Factory{})
	if err != nil {
		log.Fatalf("can't register digitalocean vpc module")
	}
}

// Factory create new vpc module.
type Factory struct{}

// New create new vpc instance.
func (f *Factory) New(providerConf digitalocean.Config, clusterState *cluster.State) (provider.Activity, error) {
	vpc := &Vpc{}
	vpc.moduleDir = filepath.Join(config.Global.ProjectRoot, "terraform/digitalocean/"+myName)
	vpc.backendConf = digitalocean.BackendSpec{
		Bucket:   providerConf.ClusterName,
		Key:      "states/terraform-" + myName + ".state",
		Endpoint: providerConf.Region + ".digitaloceanspaces.com",
	}
	vpc.config = tfVars{
		VpcID:       providerConf.Vpc,
		Region:      providerConf.Region,
		ClusterName: providerConf.ClusterName,
		VpcCIDR:     providerConf.VpcCIDR,
	}
	var err error
	vpc.terraform, err = executor.NewTerraformRunner(vpc.moduleDir, provisioner.GetAwsAuthEnv()...)
	if err != nil {
		return nil, err
	}
	vpc.terraform.LogLabels = append(vpc.terraform.LogLabels, fmt.Sprintf("cluster='%s'", providerConf.ClusterName))

	return vpc, nil
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

// Path - return module path.
func (s *Vpc) Path() string {
	return s.moduleDir
}

// Clear - remove tmp and cache files.
func (s *Vpc) Clear() error {
	return s.terraform.Clear()
}
