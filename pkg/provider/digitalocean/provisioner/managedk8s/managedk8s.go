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

// K8s class.
type K8s struct {
	providerConf digitalocean.Config
	k8sModule    provider.Activity
	state        *cluster.State
}

func init() {
	err := digitalocean.RegisterActivityFactory("provisioners", "managed-kubernetes", &Factory{})
	if err != nil {
		log.Fatalf("can't register digitalocean managed-kubernetes provisioner")
	}
}

// Factory create new digitalocean managed-kubernetes provisioner.
type Factory struct{}

// New create new instance of managed-kubernetes provisioner.
func (f *Factory) New(providerConf digitalocean.Config, state *cluster.State) (provider.Activity, error) {
	provisioner := &K8s{}
	provisioner.providerConf = providerConf
	k8sModuleFactory, exists := digitalocean.GetModulesFactories()["k8s"]
	if !exists {
		return nil, fmt.Errorf("module 'k8s', needed by provisioner is not found")
	}
	var err error
	provisioner.state = state
	provisioner.k8sModule, err = k8sModuleFactory.New(providerConf, state)
	if err != nil {
		return nil, err
	}
	return provisioner, nil
}

// Deploy managed-kubernetes provisioner modules, save kubernetes config to kubeConfig string.
// Upload kube config to s3.
func (p *K8s) Deploy() error {
	err := p.k8sModule.Deploy()
	if err != nil {
		return err
	}
	kubeConfigName := filepath.Join(p.k8sModule.Path(), "kubeconfig_"+p.providerConf.ClusterName)
	// Read kube confin from file to string.
	conf, err := ioutil.ReadFile(kubeConfigName)
	if err != nil {
		return err
	}
	if err = provisioner.CheckKubeAccess(conf); err != nil {
		return fmt.Errorf("k8s connection error: %s", err.Error())
	}
	p.state.KubeConfig = conf

	if err = provisioner.PushKubeConfig(p.providerConf.ClusterName, p.providerConf.Region, conf); err != nil {
		return fmt.Errorf("error occurred during uploading kubeconfig to spaces: %s", err.Error())
	}
	InfoTemplate := `Download and apply your kubeconfig using commands:
s3cmd get  s3://%[1]s/kubeconfig_%[1]s ~/.kube/kubeconfig_%[1]s --host-bucket='%%(bucket)s.%[2]s.digitaloceanspaces.com' --host='%[2]s.digitaloceanspaces.com'
export KUBECONFIG=~/.kube/kubeconfig_%[1]s
kubectl get ns`
	p.state.KubeAccessInfo = fmt.Sprintf(InfoTemplate, p.providerConf.ClusterName, p.providerConf.Region)
	return nil
}

// Destroy managed-kubernetes provisioner objects.
func (p *K8s) Destroy() error {
	err := p.k8sModule.Destroy()
	if err != nil {
		return err
	}
	p.state.KubeConfig = nil
	return nil
}

// Check - if s3 bucket exists.
func (p *K8s) Check() (bool, error) {
	return true, nil
}

// Path - return module path.
func (p *K8s) Path() string {
	return ""
}

// Clear - remove tmp and cache files.
func (p *K8s) Clear() error {
	return p.k8sModule.Clear()
}
