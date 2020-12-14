package reconciler

import (
	"fmt"

	"github.com/apex/log"
	"gopkg.in/yaml.v2"
)

type Backend struct {
	Name     string
	Provider string `yaml:"provider"`
	Bucket   string `yaml:"bucket"`
	Region   string `yaml:"region"`
}

func (g *Project) readBackendObj(obj map[string]interface{}) error {
	name, ok := obj["name"].(string)
	if !ok {
		return fmt.Errorf("backend object must contain field 'kind'")
	}
	spec, ok := obj["spec"]
	if !ok {
		return fmt.Errorf("backend object must contain field 'spec'")
	}
	rawSpec, err := yaml.Marshal(&spec)
	if err != nil {
		return err
	}
	backend := Backend{}
	err = yaml.Unmarshal(rawSpec, &backend)
	if err != nil {
		return err
	}
	backend.Name = name
	log.Debugf("Backend added: %+v", backend)
	g.Backends[name] = &backend
	return nil
}
