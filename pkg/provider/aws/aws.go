// Package aws - aws provider.
// Common functions of aws provider.
package aws

import (
	"fmt"

	"github.com/apex/log"

	"github.com/shalb/cluster.dev/pkg/cluster"
	"github.com/shalb/cluster.dev/pkg/provider"
	"github.com/shalb/cluster.dev/pkg/provider/aws/provisioner"
	"gopkg.in/yaml.v2"
)

// Config - aws provider config.
type Config struct {
	Region            string                 `yaml:"region"`
	Vpc               string                 `yaml:"vpc"`
	Domain            string                 `yaml:"domain"`
	Provisioner       map[string]interface{} `yaml:"provisioner"`
	ProviderType      string                 `yaml:"type"`
	AvailabilityZones []string               `yaml:"availability_zones"`
	ClusterName       string                 `yaml:"cluster_name"`
}

// Provider - main provider object.
type Provider struct {
	config Config
	state  *cluster.State
}

// BackendSpec - terraform s3 bucket backend config.
type BackendSpec struct {
	Bucket string `json:"bucket,omitempty"`
	Key    string `json:"key,omitempty"`
	Region string `json:"region,omitempty"`
}

// Register provider factory in cluster package.
func init() {
	if err := cluster.RegisterProviderFactory("aws", &Factory{}); err != nil {
		log.Fatal("can't register aws provider")
	}
}

// Factory - provider factory. Create new provider.
type Factory struct{}

// New provider, check config.
func (f *Factory) New(yamlSpec []byte, state *cluster.State, clusterName string) (cluster.Provider, error) {
	var spec Config
	err := yaml.UnmarshalStrict(yamlSpec, &spec)
	if err != nil {
		return nil, err
	}
	spec.ClusterName = clusterName
	provider := &Provider{
		config: spec,
		state:  state,
	}
	log.Debugf("Provider aws: %+v", provider)
	return provider, nil
}

// Deploy function.
func (p *Provider) Deploy() error {
	provisionerType, ok := p.config.Provisioner["type"].(string)
	if !ok {
		return fmt.Errorf("can't determinate provisioner type. Provisioner conf: %v", p.config.Provisioner)
	}
	awsDeploymentStrategy := []provider.ActivityDesc{
		{Category: "modules", Name: "backend"},
		{Category: "modules", Name: "route53"},
		{Category: "modules", Name: "vpc"},
		{Category: "provisioners", Name: provisionerType},
		{Category: "modules", Name: "addons"},
	}

	for _, opDesc := range awsDeploymentStrategy {
		opFactory, exists := providerActivitiesFactories[opDesc.Category][opDesc.Name]
		if !exists {
			return fmt.Errorf("aws provider, unknown operation '%s.%s'", opDesc.Category, opDesc.Name)
		}
		log.Infof("Deploying %s.%s", opDesc.Category, opDesc.Name)
		operation, err := opFactory.New(p.config, p.state)
		if err != nil {
			return err
		}
		defer operation.Clear()
		if err = operation.Deploy(); err != nil {
			return err
		}
	}
	return nil
}

// Destroy function.
func (p *Provider) Destroy() error {
	log.Debugf("Provider 'aws', destroying... ")
	exists, err := p.checkBackendExists()
	if err != nil {
		return err
	}
	if !exists {
		log.Info("Backend bucket does not exists, nothing to destroy.")
		return nil
	}
	provisionerType, ok := p.config.Provisioner["type"].(string)
	if !ok {
		return fmt.Errorf("can't determinate provisioner type. Provisioner conf: %v", p.config.Provisioner)
	}
	// If kubeConfig exists - destroy addons. Else - ignore addons destroying.
	p.state.KubeConfig, err = provisioner.PullKubeConfigOnce(p.config.ClusterName)
	var awsDestroyStrategy []provider.ActivityDesc
	if err == nil {
		awsDestroyStrategy = append(awsDestroyStrategy, provider.ActivityDesc{Category: "modules", Name: "addons"})
	}
	// Add provisioner and modules.
	awsDestroyStrategy = append(awsDestroyStrategy, []provider.ActivityDesc{
		{Category: "provisioners", Name: provisionerType},
		{Category: "modules", Name: "vpc"},
		{Category: "modules", Name: "route53"},
		{Category: "modules", Name: "backend"},
	}...)
	errCount := 0
	for _, opDesc := range awsDestroyStrategy {
		opFactory, exists := providerActivitiesFactories[opDesc.Category][opDesc.Name]
		if !exists {
			return fmt.Errorf("aws provider, unknown operation '%s.%s'", opDesc.Category, opDesc.Name)
		}
		log.Infof("Destroying %s.%s", opDesc.Category, opDesc.Name)
		operation, err := opFactory.New(p.config, p.state)
		if err != nil {
			return err
		}
		defer operation.Clear()

		if err = operation.Destroy(); err != nil {
			errCount++
			log.Errorf("Destroying '%s.%s' error (ignoring): %s", opDesc.Category, opDesc.Name, err.Error())
		}
	}
	if errCount > 0 {
		return fmt.Errorf("Errors occurred during the destruction. Count: %v", errCount)
	}
	return nil
}

func (p *Provider) checkBackendExists() (bool, error) {
	backendFactory, exists := providerActivitiesFactories["modules"]["backend"]
	if !exists {
		return false, fmt.Errorf("backend module is not registered in aws provider")
	}
	bk, err := backendFactory.New(p.config, p.state)
	if err != nil {
		return false, err
	}
	return bk.Check()
}
