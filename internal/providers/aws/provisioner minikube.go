package aws

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"time"

	"github.com/apex/log"
	"github.com/romanprog/c-dev/executor"
)

// ProvisionerMinikube class.
type ProvisionerMinikube struct {
	providerConf           providerConfSpec
	kubeConfig             string
	kubeConfigName         string
	kubeConfigFullFileName string
	minikubeModule         *Minikube
}

// NewProvisionerMinikube create new instance of EKS provisioner.
func NewProvisionerMinikube(providerConf providerConfSpec) (*ProvisionerMinikube, error) {
	var prv ProvisionerMinikube
	prv.providerConf = providerConf
	minikubeMod, err := NewMinikube(providerConf)
	if err != nil {
		return nil, err
	}
	// TODO check config.
	prv.minikubeModule = minikubeMod
	prv.kubeConfigName = "kubeconfig_" + providerConf.ClusterName
	prv.kubeConfigFullFileName = filepath.Join("/tmp/", prv.kubeConfigName)
	return &prv, nil
}

// Deploy EKS provisioner modules, save kubernetes config to kubeConfig string.
// Upload kube config to s3.
func (p *ProvisionerMinikube) Deploy(timeout time.Duration) error {
	err := p.minikubeModule.Deploy()
	if err != nil {
		return err
	}

	// Init bash runner in module director and export path to kubeConfig.
	varString := fmt.Sprintf("KUBECONFIG=%s", p.kubeConfigFullFileName)
	bash, err := executor.NewBashRunner("", varString)
	if err != nil {
		return err
	}
	// Ticker for pause and timeout.
	tm := time.After(timeout)
	var tick = make(<-chan time.Time)
	tick = time.Tick(5 * time.Second)
	for {
		select {
		case <-tm:
			// Timeout
			return fmt.Errorf("k8s access timeout")
		// Wait for tick.
		case <-tick:
			// Download kube config (try)
			downloadCommand := fmt.Sprintf("aws s3 cp s3://%s/%s %s", p.providerConf.ClusterName, p.kubeConfigName, p.kubeConfigFullFileName)
			stdout, stderr, err := bash.RunMutely(downloadCommand)
			if err != nil {
				log.Info("Minikube cluster is not ready yet. Will retry after 5 seconds...")
				continue
			}
			// Read kubeconfig from file.
			kubeconfig, err := ioutil.ReadFile(p.kubeConfigFullFileName)
			if err != nil {
				return err
			}
			p.kubeConfig = string(kubeconfig)
			// check k8s access
			stdout, stderr, err = bash.RunMutely("kubectl version --request-timeout=5s")
			if err == nil {
				// Connected! k8s is ready.
				log.Debugf("Kubernetes cluster is ready: %s", stdout)
				return nil
			}
			log.Info("Minikube cluster is not ready yet. Will retry after 5 seconds...")
			log.Debugf("Error check kubectl version: %s %s", stdout, stderr)
			// Connection error. Wait for next tick (try).
		}
	}
}

// Destroy minikube provisioner objects.
func (p *ProvisionerMinikube) Destroy() error {
	err := p.minikubeModule.Destroy()
	if err != nil {
		return err
	}
	p.kubeConfig = ""
	return nil
}

// GetKubeConfig return 'kubeConfig' or error if config is empty.
func (p *ProvisionerMinikube) GetKubeConfig() (string, error) {
	if p.kubeConfig == "" {
		return "", fmt.Errorf("minikube kube config is empty")
	}
	return p.kubeConfig, nil
}

// PullKubeConfig download kubeconfig from s3 and save it to this.kubeConfig variable.
func (p *ProvisionerMinikube) PullKubeConfig() error {
	bash, err := executor.NewBashRunner("")
	if err != nil {
		return err
	}
	pullCommand := fmt.Sprintf("aws s3 cp s3://%s/%s %s", p.providerConf.ClusterName, p.kubeConfigName, p.kubeConfigFullFileName)
	stdout, stderr, err := bash.RunMutely(pullCommand)
	if err != nil {
		log.Debugf("aws s3 cp output:\n%s\nError:\n%s", stdout, stderr)
		return err
	}
	log.Debugf("Kubeconfig pulled. aws s3 cp output: \n%s", stdout)
	kubeConfig, err := ioutil.ReadFile(p.kubeConfigFullFileName)
	if err != nil {
		return err
	}
	p.kubeConfig = string(kubeConfig)

	return nil
}
