package aws

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/apex/log"
	"github.com/romanprog/c-dev/executor"
)

// ProvisionerEks class.
type ProvisionerEks struct {
	providerConf           providerConfSpec
	kubeConfig             string
	kubeConfigName         string
	kubeConfigFullFileName string
	eksModule              *Eks
}

// NewProvisionerEks create new instance of EKS provisioner.
func NewProvisionerEks(providerConf providerConfSpec) (*ProvisionerEks, error) {
	var prv ProvisionerEks
	prv.providerConf = providerConf
	eksMod, err := NewEks(providerConf)
	if err != nil {
		return nil, err
	}
	// TODO check config.
	prv.eksModule = eksMod
	prv.kubeConfigName = "kubeconfig_" + providerConf.ClusterName
	prv.kubeConfigFullFileName = filepath.Join("/tmp", prv.kubeConfigName)
	return &prv, nil
}

// Deploy EKS provisioner modules, save kubernetes config to kubeConfig string.
// Upload kube config to s3.
func (p *ProvisionerEks) Deploy(timeout time.Duration) error {
	err := p.eksModule.Deploy()
	if err != nil {
		return err
	}
	// Read kube confin from file to string.
	conf, err := ioutil.ReadFile(p.kubeConfigName)
	if err != nil {
		return err
	}
	p.kubeConfig = string(conf)

	// Write kubeconfig to standart path.
	err = ioutil.WriteFile(p.kubeConfigFullFileName, conf, os.ModePerm)
	if err != nil {
		return err
	}
	// Init bash runner in module director and export path to kubeConfig.
	varString := fmt.Sprintf("KUBECONFIG=%s", p.kubeConfigFullFileName)
	bash, err := executor.NewBashRunner("", varString)
	if err != nil {
		return err
	}
	// Ticker for pause.
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
			// Test k8s access.
			stdout, stderr, err := bash.RunMutely("kubectl version --request-timeout=5s")
			if err == nil {
				// Connected! k8s is ready.
				// Upload kubeConfig to s3
				uploadCommand := fmt.Sprintf("aws s3 cp %s s3://%s/%s", p.kubeConfigFullFileName, p.providerConf.ClusterName, p.kubeConfigName)
				return bash.Run(uploadCommand)
			}
			log.Debugf("%s %s", stdout, stderr)
			// Connection error. Wait for next tick (try).
		}
	}
}

// Destroy EKS provisioner objects.
func (p *ProvisionerEks) Destroy() error {
	err := p.eksModule.Destroy()
	if err != nil {
		return err
	}
	p.kubeConfig = ""
	return nil
}

// GetKubeConfig return 'kubeConfig' or error if config is empty.
func (p *ProvisionerEks) GetKubeConfig() (string, error) {
	if p.kubeConfig == "" {
		return "", fmt.Errorf("eks kube config is empty")
	}
	return p.kubeConfig, nil
}

// PullKubeConfig download kubeconfig from s3 and save it to this.kubeConfig variable.
func (p *ProvisionerEks) PullKubeConfig() error {
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
