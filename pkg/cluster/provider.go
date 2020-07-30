package cluster

import (
	"fmt"
)

type Provider interface {
	Deploy() error
	Destroy() error
}

type ProviderFactory interface {
	NewProvider(cfg []byte) (Provider, error)
}

var providerFactories = map[string]ProviderFactory{}

// Each Provider registers itself so it can be used when specified in a Manifest
func RegisterProviderFactory(providerType string, factory ProviderFactory) error {
	if _, exists := providerFactories[providerType]; exists {
		return fmt.Errorf("provider %v is already registered", providerType)
	}
	providerFactories[providerType] = factory
	return nil
}

func NewProvider(providerType string, cfg []byte) (Provider, error) {
	factory, exists := providerFactories[providerType]
	if !exists {
		return nil, fmt.Errorf("provider %v is not registered", providerType)
	}
	p, err := factory.NewProvider(cfg)
	if err != nil {
		return nil, err
	}
	return p, nil
}
