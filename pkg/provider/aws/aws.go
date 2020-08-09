// Package aws - aws provider.
// Common functions of aws provider.
package aws

import (
	"fmt"

	"github.com/apex/log"

	"github.com/shalb/cluster.dev/pkg/cluster"
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

// Operation common interface for module and provisioner.
type Operation interface {
	// Apply module with his defined configuration.
	Deploy() error
	// Destroy infrastructure, created by module.
	Destroy() error
	// Some modules checks.
	Check() (bool, error)
	// Path to directory with module files.
	Path() string
	// Clear module tmp files and cache.
	Clear() error
}

// OperationFactory common interface for modules and provisioners factories.
type OperationFactory interface {
	New(Config, *cluster.State) (Operation, error)
}

// Base of provider operations (modules and provisioners).
var providerOperationsFactories = map[string]map[string]OperationFactory{
	"modules":      make(map[string]OperationFactory),
	"provisioners": make(map[string]OperationFactory),
}

type operationDesc struct {
	class string
	name  string
}

// Deploy function.
func (p *Provider) Deploy() error {
	provisionerType, ok := p.config.Provisioner["type"].(string)
	if !ok {
		return fmt.Errorf("can't determinate provisioner type. Provisioner conf: %v", p.config.Provisioner)
	}
	awsDeploymentStrategy := []operationDesc{
		{class: "modules", name: "backend"},
		{class: "modules", name: "route53"},
		{class: "modules", name: "vpc"},
		{class: "provisioners", name: provisionerType},
		{class: "modules", name: "addons"},
	}

	for _, opDesc := range awsDeploymentStrategy {
		opFactory, exists := providerOperationsFactories[opDesc.class][opDesc.name]
		if !exists {
			return fmt.Errorf("aws provider, unknown operation '%s.%s'", opDesc.class, opDesc.name)
		}
		log.Infof("Deploying %s.%s", opDesc.class, opDesc.name)
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
	var awsDestroyStrategy []operationDesc
	if err == nil {
		awsDestroyStrategy = append(awsDestroyStrategy, operationDesc{class: "modules", name: "addons"})
	}
	// Add provisioner and modules.
	awsDestroyStrategy = append(awsDestroyStrategy, []operationDesc{
		{class: "provisioners", name: provisionerType},
		{class: "modules", name: "vpc"},
		{class: "modules", name: "route53"},
		{class: "modules", name: "backend"},
	}...)
	errCount := 0
	for _, opDesc := range awsDestroyStrategy {
		opFactory, exists := providerOperationsFactories[opDesc.class][opDesc.name]
		if !exists {
			return fmt.Errorf("aws provider, unknown operation '%s.%s'", opDesc.class, opDesc.name)
		}
		log.Infof("Destroying %s.%s", opDesc.class, opDesc.name)
		operation, err := opFactory.New(p.config, p.state)
		if err != nil {
			return err
		}
		defer operation.Clear()

		if err = operation.Destroy(); err != nil {
			errCount++
			log.Errorf("Destroying '%s.%s' error (ignoring): %s")
		}
	}
	if errCount > 0 {
		return fmt.Errorf("Errors occurred during the destruction. Count: %v", errCount)
	}
	return nil
}

func (p *Provider) checkBackendExists() (bool, error) {
	log.Debugf("Provider 'aws', destroying... ")
	backendFactory, exists := providerOperationsFactories["modules"]["backend"]
	if !exists {
		return false, fmt.Errorf("backend module is not registered in aws provider")
	}
	bk, err := backendFactory.New(p.config, p.state)
	if err != nil {
		return false, err
	}
	return bk.Check()
}
