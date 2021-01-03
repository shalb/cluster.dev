package project

import (
	"fmt"

	"github.com/apex/log"
	"gopkg.in/yaml.v2"
)

const backendObjKindKey = "backend"

// Backend interface for backend provider.
type Backend interface {
	Name() string
	Provider() string
	GetBackendHCL(Module) ([]byte, error)
	GetRemoteStateHCL(Module) ([]byte, error)
}

// BackendsFactory - interface for backend provider factory. New() creates backend.
type BackendsFactory interface {
	New([]byte, string) (Backend, error)
}

// RegisterBackendFactory - register factory of some provider (like s3) in map.
func RegisterBackendFactory(f BackendsFactory, provider string) error {
	if _, exists := BackendsFactories[provider]; exists {
		return fmt.Errorf("backend with provider name '%v' already exists", provider)
	}
	BackendsFactories[provider] = f
	return nil
}

// BackendsFactories map of backend providers factories. Use BackendsFactories["prov_name"].New() to create backend of provider 'prov_name'
var BackendsFactories = map[string]BackendsFactory{}

func (g *Project) readBackendObj(obj map[string]interface{}) error {
	name, ok := obj["name"].(string)
	if !ok {
		return fmt.Errorf("backend object must contain field 'name'")
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
	log.Infof("Backend '%v' added", name)
	return nil
}
