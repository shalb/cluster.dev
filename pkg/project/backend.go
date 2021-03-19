package project

import (
	"fmt"

	"github.com/apex/log"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"gopkg.in/yaml.v3"
)

const backendObjKindKey = "backend"

// Backend interface for backend provider.
type Backend interface {
	Name() string
	Provider() string
	GetBackendHCL(string, string) (*hclwrite.File, error)
	GetBackendBytes(string, string) ([]byte, error)
	GetRemoteStateHCL(string, string) ([]byte, error)
	State() map[string]interface{}
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

func (p *Project) readBackends() error {
	// Read and parse backends.
	bks, exists := p.objects[backendObjKindKey]
	if !exists {
		err := fmt.Errorf("no backend found, at least one backend needed")
		log.Debug(err.Error())
		return err
	}
	for _, bk := range bks {
		err := p.readBackendObj(bk)
		if err != nil {
			return fmt.Errorf("loading backend: %v", err.Error())
		}
	}
	return nil
}

func (p *Project) readBackendObj(obj ObjectData) error {
	name, ok := obj.data["name"].(string)
	if !ok {
		return fmt.Errorf("backend object must contain field 'name'")
	}
	spec, ok := obj.data["spec"]
	if !ok {
		return fmt.Errorf("backend object must contain field 'spec'")
	}
	provider, ok := obj.data["provider"].(string)
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
	p.Backends[name] = b
	log.Infof("Backend '%v' added", name)
	return nil
}
