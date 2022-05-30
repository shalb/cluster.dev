package project

import (
	"fmt"

	"github.com/apex/log"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"gopkg.in/yaml.v3"
)

const backendObjKindKey = "Backend"
const defaultLocalBackendConfig = `
name: default
kind: Backend
provider: local
spec:
  path: ".cluster.dev/states/"
`

// Backend interface for backend provider.
type Backend interface {
	Name() string
	Provider() string
	GetBackendHCL(string, string) (*hclwrite.File, error)
	GetBackendBytes(string, string) ([]byte, error)
	GetRemoteStateHCL(string, string) ([]byte, error)
	LockState() error
	UnlockState() error
	WriteState(stateData string) error
	ReadState() (string, error)
}

// BackendsFactory - interface for backend provider factory. New() creates backend.
type BackendsFactory interface {
	New([]byte, string, *Project) (Backend, error)
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
	if exists {
		for _, bk := range bks {
			err := p.readBackendObj(bk)
			if err != nil {
				return fmt.Errorf("reading backend: %v", err.Error())
			}
		}
	}
	return addDefaultBackend(p)
}

func (p *Project) readBackendObj(obj ObjectData) error {
	name, ok := obj.data["name"].(string)
	if !ok {
		return fmt.Errorf("config must contain field 'name'")
	}
	spec, ok := obj.data["spec"]
	if !ok {
		return fmt.Errorf("'%v': config must contain field 'spec'", name)
	}
	provider, ok := obj.data["provider"].(string)
	if !ok {
		return fmt.Errorf("'%v': must contain field 'provider'", name)
	}
	rawSpec, err := yaml.Marshal(&spec)
	if err != nil {
		return err
	}
	factory, exists := BackendsFactories[provider]
	if !exists {
		return fmt.Errorf("'%v': provider does not found: %v", name, provider)
	}
	b, err := factory.New(rawSpec, name, p)
	if err != nil {
		return err
	}
	p.Backends[name] = b
	log.Debugf("Backend added: %v", name)
	return nil
}

func addDefaultBackend(p *Project) error {
	if _, exists := p.Backends["default"]; exists {
		return fmt.Errorf("read backends: name 'default' is reserved, use another backend name")
	}
	defBkParsed := map[string]interface{}{}
	err := yaml.Unmarshal([]byte(defaultLocalBackendConfig), &defBkParsed)
	if err != nil {
		return err
	}
	obj := ObjectData{
		filename: "",
		data:     defBkParsed,
	}

	return p.readBackendObj(obj)
}
