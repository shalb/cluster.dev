package domain

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

const myName = "domain"

// Variables set for route53 module tfvars.
type domainVarsSpec struct {
	Region         string `json:"region"`
	ClusterName    string `json:"cluster_name"`
	ClusterDomain  string `json:"cluster_domain"`
	ZoneDelegation string `json:"zone_delegation"`
}

// Domain type for domain module instance.
type Domain struct {
	config      domainVarsSpec
	backendConf digitalocean.BackendSpec
	terraform   *executor.TerraformRunner
	backendKey  string
	moduleDir   string
}

func init() {
	err := digitalocean.RegisterActivityFactory("modules", myName, &Factory{})
	if err != nil {
		log.Fatalf("can't register digitalocean " + myName + " module")
	}
}

// Factory create new route53 module.
type Factory struct{}

// New create new domain instance.
func (f *Factory) New(providerConf digitalocean.Config, clusterState *cluster.State) (provider.Activity, error) {
	route53 := &Domain{}
	route53.moduleDir = filepath.Join(config.Global.ProjectRoot, "terraform/digitalocean/"+myName)
	route53.backendConf = digitalocean.BackendSpec{
		Bucket:   providerConf.ClusterName,
		Key:      "states/terraform-dns.state",
		Endpoint: providerConf.Region + ".digitaloceanspaces.com",
	}
	zoneDelegation := "false"
	if providerConf.Domain == "cluster.dev" {
		zoneDelegation = "true"
	}
	route53.config = domainVarsSpec{
		Region:         providerConf.Region,
		ClusterName:    providerConf.ClusterName,
		ClusterDomain:  providerConf.Domain,
		ZoneDelegation: zoneDelegation,
	}
	var err error
	route53.terraform, err = executor.NewTerraformRunner(route53.moduleDir, provisioner.GetAwsAuthEnv()...)
	if err != nil {
		return nil, err
	}
	route53.terraform.LogLabels = append(route53.terraform.LogLabels, fmt.Sprintf("cluster='%s'", providerConf.ClusterName))
	return route53, nil
}

// Deploy - create domain.
func (r *Domain) Deploy() error {
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

// Destroy - remove domain.
func (r *Domain) Destroy() error {
	// Init terraform without backend spec.
	err := r.terraform.Init(r.backendConf)
	if err != nil {
		return err
	}
	// Plan.
	return r.terraform.Destroy(r.config, "-compact-warnings")
}

// Check - if s3 bucket exists.
func (r *Domain) Check() (bool, error) {
	return true, nil
}

// Path - return module path.
func (r *Domain) Path() string {
	return r.moduleDir
}

// Clear - remove tmp and cache files.
func (r *Domain) Clear() error {
	return r.terraform.Clear()
}
