package minikube

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/shalb/cluster.dev/pkg/cluster"
	"gopkg.in/yaml.v3"
	"log"
)

// register provisioner factory in cluster package
func init() {
	err := cluster.RegisterProvisionerFactory("aws", "minikube", &Factory{})
	if err != nil {
		log.Fatal("can't register minikube provisioner")
	}
}

type Factory struct{}

func (f *Factory) NewProvisioner(cfg []byte) (cluster.Provisioner, error) {
	p := &Minikube{}
	if err := yaml.Unmarshal(cfg, p); err != nil {
		return nil, fmt.Errorf("error occured during YAML unmarshalling %v", err)
	}
	if err := Validate(p); err != nil {
		return nil, err
	}

	return p, nil
}

func Validate(p *Minikube) error {
	v := validator.New()
	err := v.Struct(p)

	if err != nil {
		return err.(validator.ValidationErrors)
	}

	return nil
}

type Minikube struct {
	InstanceType string `yaml:"instance_type" validate:"required"`
}
