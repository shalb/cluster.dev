package addons

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/apex/log"
	"github.com/shalb/cluster.dev/internal/config"
	"github.com/shalb/cluster.dev/internal/executor"
	"github.com/shalb/cluster.dev/pkg/cluster"
	"github.com/shalb/cluster.dev/pkg/provider"
	"github.com/shalb/cluster.dev/pkg/provider/digitalocean"
	"github.com/shalb/cluster.dev/pkg/provider/digitalocean/provisioner"
)

const moduleName = "addons"

// Variables set for module tfvars.
type tfVars struct {
	Region             string `json:"region"`
	ClusterName        string `json:"cluster_name"`
	ConfigPath         string `json:"config_path"`
	ClusterCloudDomain string `json:"cluster_cloud_domain"`
	DoToken            string `json:"do_token"`
}

// Addons - type for module instance.
type Addons struct {
	config      tfVars
	backendConf digitalocean.BackendSpec
	terraform   *executor.TerraformRunner
	state       *cluster.State
	kubeConfig  []byte
	backendKey  string
	moduleDir   string
	tmpDir      string
}

func init() {
	err := digitalocean.RegisterActivityFactory("modules", moduleName, &Factory{})
	if err != nil {
		log.Fatalf("can't register digitalocean addons module")
	}
}

// Factory create new addons module.
type Factory struct{}

// New create new addons instance.
func (f *Factory) New(providerConf digitalocean.Config, clusterState *cluster.State) (provider.Activity, error) {
	addons := &Addons{}
	// Module dir.
	addons.moduleDir = filepath.Join(config.Global.ProjectRoot, "terraform/digitalocean/"+moduleName)
	// Module state name.
	addons.backendKey = "states/terraform-addons.state"
	// Set backend config.
	addons.backendConf = digitalocean.BackendSpec{
		Bucket:   providerConf.ClusterName,
		Key:      addons.backendKey,
		Endpoint: providerConf.Region + ".digitaloceanspaces.com",
	}
	// Set tfVars.
	addons.config.ClusterName = providerConf.ClusterName
	addons.config.Region = providerConf.Region
	addons.config.ClusterCloudDomain = providerConf.Domain
	addons.kubeConfig = clusterState.KubeConfig
	addons.state = clusterState
	addons.config.DoToken = os.Getenv("DIGITALOCEAN_TOKEN")
	if addons.config.DoToken == "" {
		return nil, fmt.Errorf("env variable DIGITALOCEAN_TOKEN is not set")
	}
	var err error
	addons.tmpDir, err = ioutil.TempDir("", "cluster-dev-addons-*")
	if err != nil {
		return nil, fmt.Errorf("can't create tmp dir: %s", err.Error())
	}
	// Save cube config to tmp file.
	addons.config.ConfigPath = filepath.Join(addons.tmpDir, "kube_config")

	// Write kube config to file.
	err = ioutil.WriteFile(addons.config.ConfigPath, clusterState.KubeConfig, os.ModePerm)
	if err != nil {
		return nil, err
	}

	// Init terraform runner in module directory.
	addons.terraform, err = executor.NewTerraformRunner(addons.moduleDir, provisioner.GetAwsAuthEnv()...)
	if err != nil {
		return nil, err
	}
	addons.terraform.LogLabels = append(addons.terraform.LogLabels, fmt.Sprintf("cluster='%s'", providerConf.ClusterName))
	return addons, nil
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
	InfoTemplate := `Download and apply your kubeconfig using commands:
s3cmd get s3://%[1]s/addons/auth.yaml ~/auth.yaml --host-bucket='%%(bucket)s.%[2]s.digitaloceanspaces.com' --host='%[2]s.digitaloceanspaces.com'
cat ~/auth.yaml`
	s.state.AddonsAccessInfo = fmt.Sprintf(InfoTemplate, s.config.ClusterName, s.config.Region)
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

// Path - return terraform module path.
func (s *Addons) Path() string {
	return s.moduleDir
}

// Clear - remove tmp and cache files.
func (s *Addons) Clear() error {
	err := os.RemoveAll(s.tmpDir)
	if err != nil {
		log.Debugf("Addons clear error (igniring): %s", err.Error())
	}
	return s.terraform.Clear()
}
