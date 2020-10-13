package cluster

import (
	"fmt"

	"gopkg.in/yaml.v2"
)

// Provider - interface fo all providers.
type Provider interface {
	Deploy() error
	Destroy() error
}

// ProviderFactory - interface for providers factories.
type ProviderFactory interface {
	New(cfg []byte, state *State, clusterName string) (Provider, error)
}

var providerFactories = map[string]ProviderFactory{}

// RegisterProviderFactory - each Provider registers itself so it can be used when specified in a Manifest
func RegisterProviderFactory(providerType string, factory ProviderFactory) error {
	if _, exists := providerFactories[providerType]; exists {
		return fmt.Errorf("provider %v is already registered", providerType)
	}
	providerFactories[providerType] = factory
	return nil
}

// NewProvider return new provider instance of providerType.
func NewProvider(clusterConfig Config, state *State) (Provider, error) {
	providerCfg, err := getProviderConfig(clusterConfig)
	if err != nil {
		return nil, err
	}
	providerType, exists := clusterConfig.ProviderConfig.(map[interface{}]interface{})["type"].(string)
	if !exists {
		return nil, fmt.Errorf("YAML must contain provider.type field")
	}
	factory, exists := providerFactories[providerType]
	if !exists {
		return nil, fmt.Errorf("provider %v is not registered", providerType)
	}
	p, err := factory.New(providerCfg, state, clusterConfig.ClusterFullName)
	if err != nil {
		return nil, err
	}
	return p, nil
}

// getProviderConfig parse cluster config check if provider category exists. Marshal it and return in raw yaml.
func getProviderConfig(clusterConfig Config) ([]byte, error) {

	if clusterConfig.ProviderConfig == nil {
		return nil, fmt.Errorf("provider config is empty")
	}
	// Convert provider subcategory to raw yaml.
	providerData, err := yaml.Marshal(clusterConfig.ProviderConfig)
	if err != nil {
		return nil, err
	}
	return providerData, nil
}
