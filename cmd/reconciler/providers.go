package main

import (
	"fmt"

	"github.com/shalb/cluster.dev/internal/providers/aws"

	"github.com/apex/log"
	"gopkg.in/yaml.v2"
)

// ProviderCommon - interface for all providers.
type ProviderCommon interface {
	Init(providerSpec []byte, clusterName string) error
	Deploy() error
	Destroy() error
	GetKubeConfig() string
}

var providers map[string]ProviderCommon

func providersLoad() error {
	providers = make(map[string]ProviderCommon)

	// Add AWS provider to providers map.
	awsProvider := aws.Provider{}
	providers["aws"] = &awsProvider
	return nil
	// Add other providers.
	// ...

}

func newProvider(spec []byte, clusterName string) (ProviderCommon, error) {
	// Parse provider spec.
	var provConf struct {
		ProvType string `yaml:"type"`
	}
	log.Debug("Unmarshal yaml")
	err := yaml.Unmarshal(spec, &provConf)
	if err != nil {
		return nil, err
	}
	// Check provider type exists.
	if provConf.ProvType == "" {
		return nil, fmt.Errorf("undefined provider type")
	}

	log.Debug("Check provider")
	prov, ok := providers[provConf.ProvType]
	if !ok {
		return nil, fmt.Errorf("provider '%s' not found", provConf.ProvType)
	}

	log.Debug("Init provider with config")
	err = prov.Init(spec, clusterName)
	if err != nil {
		return nil, err
	}
	return prov, nil
}
