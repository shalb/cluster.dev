package aws

import (
	"fmt"
	"github.com/shalb/cluster.dev/pkg/cluster"
	"gopkg.in/go-playground/validator.v10"
	"gopkg.in/yaml.v3"
	"log"
)

// register provider factory in cluster package
func init() {
	err := cluster.RegisterProviderFactory("aws", &Factory{})
	if err != nil {
		log.Fatal("can't register aws provider")
	}
}

type Factory struct{}

func (f *Factory) NewProvider(cfg []byte) (cluster.Provider, error) {
	// Set default values
	p := &Aws{
		Vpc:    "default",
		Domain: "cluster.dev",
	}
	if err := yaml.Unmarshal(cfg, p); err != nil {
		return nil, fmt.Errorf("error occured during YAML unmarshalling %v", err)
	}

	if err := Validate(p); err != nil {
		return nil, err
	}

	provisionerType, exists := p.Provisioner.(map[string]interface{})["type"].(string)
	if !exists {
		return nil, fmt.Errorf("YAML must contain provisioner.type field")
	}

	provisionerYaml, err := yaml.Marshal(p.Provisioner)
	if err != nil {
		return nil, err
	}

	provisioner, err := cluster.NewProvisioner("aws", provisionerType, provisionerYaml)
	if err != nil {
		return nil, err
	}
	p.Provisioner = provisioner

	return p, nil
}

func Validate(p *Aws) error {
	v := validator.New()
	err := v.Struct(p)

	if err != nil {
		return err.(validator.ValidationErrors)
	}

	return nil
}

type Aws struct {
	Type              string      `yaml:"type" validate:"required"`
	Region            string      `yaml:"region" validate:"required"`
	AvailabilityZones []string    `yaml:"availability_zones"`
	Vpc               string      `yaml:"vpc"`
	Domain            string      `yaml:"domain"`
	Provisioner       interface{} `yaml:"provisioner" validate:"required"`
}

func (a *Aws) Deploy() error {
	return nil
}

func (a *Aws) Destroy() error {
	return nil
}
