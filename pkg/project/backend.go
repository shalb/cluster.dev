package project

import (
	"fmt"

	"gopkg.in/yaml.v2"
)

type Backend interface {
	Name() string
	Provider() string
	GetBackendHCL(Module) ([]byte, error)
	GetRemoteStateHCL(Module) ([]byte, error)
}

type BackendsFactory interface {
	New([]byte, string) (Backend, error)
}

func RegisterBackendFactory(f BackendsFactory, provider string) error {
	if _, exists := BackendsFactories[provider]; exists {
		return fmt.Errorf("backend with provider name '%v' already exists", provider)
	}
	BackendsFactories[provider] = f
	return nil
}

var BackendsFactories = map[string]BackendsFactory{}

func (g *Project) readBackendObj(obj map[string]interface{}) error {
	name, ok := obj["name"].(string)
	if !ok {
		return fmt.Errorf("backend object must contain field 'kind'")
	}
	spec, ok := obj["spec"]
	if !ok {
		return fmt.Errorf("backend object must contain field 'spec'")
	}
	provider, ok := obj["provider"].(string)
	if !ok {
		return fmt.Errorf("backend object must contain field 'provider'")
	}
	rawSpec, err := yaml.Marshal(&spec)
	if err != nil {
		return err
	}
	factory, exists := BackendsFactories[provider]
	if !exists {
		return fmt.Errorf("backend provider '%s' not found", provider)
	}
	b, err := factory.New(rawSpec, name)
	if err != nil {
		return err
	}
	g.Backends[name] = b
	return nil
}
