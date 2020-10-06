package k8s

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
	"gopkg.in/yaml.v2"
)

// Variables set for k8s module tfvars.
type tfVars struct {
	Region       string `json:"region"`
	ClusterName  string `json:"cluster_name"`
	NodeType     string `json:"node_type" yaml:"nodeSize"`
	K8sVersion   string `json:"k8s_version" yaml:"version"`
	MinNodeCount int    `json:"min_node_count" yaml:"minNodes"`
	MaxNodeCount int    `json:"max_node_count" yaml:"maxNodes"`
}

const myName = "k8s"

// K8s - type for k8s module.
type K8s struct {
	config      tfVars
	backendConf digitalocean.BackendSpec
	terraform   *executor.TerraformRunner
	backendKey  string
	moduleDir   string
}

func init() {
	err := digitalocean.RegisterActivityFactory("modules", "k8s", &Factory{})
	if err != nil {
		log.Fatalf("can't register digitalocean k8s module")
	}
}

// Factory create new k8s module.
type Factory struct{}

// New create new k8s instance.
func (f *Factory) New(providerConf digitalocean.Config, clusterState *cluster.State) (provider.Activity, error) {
	k8s := &K8s{}
	k8s.moduleDir = filepath.Join(config.Global.ProjectRoot, "terraform/digitalocean/"+myName)
	k8s.backendKey = "states/terraform-" + myName + ".state"
	k8s.backendConf = digitalocean.BackendSpec{
		Bucket:   providerConf.ClusterName,
		Key:      k8s.backendKey,
		Endpoint: providerConf.Region + ".digitaloceanspaces.com",
	}
	rawProvisionerData, err := yaml.Marshal(providerConf.Provisioner)
	if err != nil {
		return nil, fmt.Errorf("error occurret while marshal provisioner config: %s", err.Error())
	}
	if err = yaml.Unmarshal(rawProvisionerData, &k8s.config); err != nil {
		return nil, fmt.Errorf("error occurret while parsing provisioner config: %s", err.Error())
	}

	k8s.config.ClusterName = providerConf.ClusterName
	k8s.config.Region = providerConf.Region

	k8s.terraform, err = executor.NewTerraformRunner(k8s.moduleDir, provisioner.GetAwsAuthEnv()...)
	if err != nil {
		return nil, err
	}
	k8s.terraform.LogLabels = append(k8s.terraform.LogLabels, fmt.Sprintf("cluster='%s'", providerConf.ClusterName))
	return k8s, nil
}

// Deploy k8s.
func (s *K8s) Deploy() error {
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

// Destroy - remove k8s.
func (s *K8s) Destroy() error {
	// Init terraform.
	err := s.terraform.Init(s.backendConf)
	if err != nil {
		return err
	}
	// Destroy.
	return s.terraform.Destroy(s.config, "-compact-warnings")
}

// Check - do nothing, not used yet.
func (s *K8s) Check() (bool, error) {
	return true, nil
}

// Path - return terraform module path.
func (s *K8s) Path() string {
	return s.moduleDir
}

// Clear - remove tmp and cache files.
func (s *K8s) Clear() error {
	return s.terraform.Clear()
}
