package managedk8s

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/apex/log"
	"github.com/shalb/cluster.dev/pkg/cluster"
	"github.com/shalb/cluster.dev/pkg/provider"
	"github.com/shalb/cluster.dev/pkg/provider/digitalocean"
	"github.com/shalb/cluster.dev/pkg/provider/digitalocean/provisioner"
)

// Eks class.
type Eks struct {
	providerConf digitalocean.Config
	eksModule    provider.Activity
	state        *cluster.State
}

func init() {
	err := digitalocean.RegisterActivityFactory("provisioners", "managed-kubernetes", &Factory{})
	if err != nil {
		log.Fatalf("can't register digitalocean eks provisioner")
	}
}

// Factory create new digitalocean eks provisioner.
type Factory struct{}

// New create new instance of EKS provisioner.
func (f *Factory) New(providerConf digitalocean.Config, state *cluster.State) (provider.Activity, error) {
	provisioner := &Eks{}
	provisioner.providerConf = providerConf
	eksModuleFactory, exists := digitalocean.GetModulesFactories()["eks"]
	if !exists {
		return nil, fmt.Errorf("module 'eks', needed by provisioner is not found")
	}
	var err error
	provisioner.state = state
	provisioner.eksModule, err = eksModuleFactory.New(providerConf, state)
	if err != nil {
		return nil, err
	}
	return provisioner, nil
}

// Deploy EKS provisioner modules, save kubernetes config to kubeConfig string.
// Upload kube config to s3.
func (p *Eks) Deploy() error {
	err := p.eksModule.Deploy()
	if err != nil {
		return err
	}
	kubeConfigName := filepath.Join(p.eksModule.Path(), "kubeconfig_"+p.providerConf.ClusterName)
	// Read kube confin from file to string.
	conf, err := ioutil.ReadFile(kubeConfigName)
	if err != nil {
		return err
	}
	if err = provisioner.CheckKubeAccess(conf); err != nil {
		return fmt.Errorf("k8s connection error: %s", err.Error())
	}
	p.state.KubeConfig = conf

	if err = provisioner.PushKubeConfig(p.providerConf.ClusterName, conf); err != nil {
		return fmt.Errorf("error occurred during uploading kubeconfig to s3 bucket: %s", err.Error())
	}
	return nil
}

// Destroy EKS provisioner objects.
func (p *Eks) Destroy() error {
	err := p.eksModule.Destroy()
	if err != nil {
		return err
	}
	p.state.KubeConfig = nil
	return nil
}

// Check - if s3 bucket exists.
func (p *Eks) Check() (bool, error) {
	return true, nil
}

// Path - return module path.
func (p *Eks) Path() string {
	return ""
}

// Clear - remove tmp and cache files.
func (p *Eks) Clear() error {
	return p.eksModule.Clear()
}
