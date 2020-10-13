package route53

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

// Variables set for route53 module tfvars.
type route53VarsSpec struct {
	Region         string `json:"region"`
	ClusterName    string `json:"cluster_name"`
	ClusterDomain  string `json:"cluster_domain"`
	ZoneDelegation string `json:"zone_delegation"`
}

// Route53 type for route53 module instance.
type Route53 struct {
	config      route53VarsSpec
	backendConf aws.BackendSpec
	terraform   *executor.TerraformRunner
	backendKey  string
	moduleDir   string
}

func init() {
	err := aws.RegisterActivityFactory("modules", "route53", &Factory{})
	if err != nil {
		log.Fatalf("can't register aws route53 module")
	}
}

// Factory create new route53 module.
type Factory struct{}

// New create new eks instance.
func (f *Factory) New(providerConf aws.Config, clusterState *cluster.State) (provider.Activity, error) {
	route53 := &Route53{}
	route53.moduleDir = filepath.Join(config.Global.ProjectRoot, "terraform/aws/route53")
	route53.backendConf = aws.BackendSpec{
		Bucket: providerConf.ClusterName,
		Key:    "states/terraform-dns.state",
		Region: providerConf.Region,
	}
	zoneDelegation := "false"
	if providerConf.Domain == "cluster.dev" {
		zoneDelegation = "true"
	}
	route53.config = route53VarsSpec{
		Region:         providerConf.Region,
		ClusterName:    providerConf.ClusterName,
		ClusterDomain:  providerConf.Domain,
		ZoneDelegation: zoneDelegation,
	}
	var err error
	route53.terraform, err = executor.NewTerraformRunner(route53.moduleDir)
	if err != nil {
		return nil, err
	}
	route53.terraform.LogLabels = append(route53.terraform.LogLabels, fmt.Sprintf("cluster='%s'", providerConf.ClusterName))
	return route53, nil
}

// Deploy - create route53.
func (r *Route53) Deploy() error {
	// sss
	log.Debug("Terraform init/plan.")
	err := r.terraform.Clear()
	if err != nil {
		return err
	}
	// Init terraform without backend speck.
	err = r.terraform.Init(r.backendConf)
	if err != nil {
		return err
	}
	// Plan.
	err = r.terraform.Plan(r.config, "-compact-warnings", "-out=tfplan")
	if err != nil {
		return err
	}
	// Apply. Create DNS.
	err = r.terraform.ApplyPlan("tfplan", "-compact-warnings")
	if err != nil {
		return err
	}
	return nil
}

// Destroy - remove s3 bucket.
func (r *Route53) Destroy() error {
	// Init terraform without backend speck.
	err := r.terraform.Init(r.backendConf)
	if err != nil {
		return err
	}
	// Plan.
	return r.terraform.Destroy(r.config, "-compact-warnings")
}

// Check - if s3 bucket exists.
func (r *Route53) Check() (bool, error) {
	return true, nil
}

// Path - return module path.
func (r *Route53) Path() string {
	return r.moduleDir
}

// Clear - remove tmp and cache files.
func (r *Route53) Clear() error {
	return r.terraform.Clear()
}
