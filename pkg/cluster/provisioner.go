package cluster

import (
	"fmt"
)

type Provisioner interface {
	//Deploy() error
	//Destroy() error
}

type ProvisionerFactory interface {
	NewProvisioner(cfg []byte) (Provisioner, error)
}

var provisionerFactories = map[string]map[string]ProvisionerFactory{}

// Each Provider registers itself so it can be used when specified in a Manifest
func RegisterProvisionerFactory(providerType string, provisionerType string, factory ProvisionerFactory) error {
	if _, exists := provisionerFactories[providerType]; !exists {
		provisionerFactories[providerType] = map[string]ProvisionerFactory{}
	}
	if _, exists := provisionerFactories[providerType][provisionerType]; exists {
		return fmt.Errorf("provider %v is already registered", provisionerType)
	}
	provisionerFactories[providerType][provisionerType] = factory
	return nil
}

func NewProvisioner(providerType string, provisionerType string, cfg []byte) (Provisioner, error) {
	if _, exists := provisionerFactories[providerType]; !exists {
		return nil, fmt.Errorf("no provisioners registered for provider %v", providerType)
	}
	factory, exists := provisionerFactories[providerType][provisionerType]
	if !exists {
		return nil, fmt.Errorf("provisioner %v is not registered", provisionerType)
	}
	p, err := factory.NewProvisioner(cfg)
	if err != nil {
		return nil, err
	}
	return p, nil
}
