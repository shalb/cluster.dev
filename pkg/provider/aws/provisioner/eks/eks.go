package eks

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/shalb/cluster.dev/pkg/cluster"
	"gopkg.in/yaml.v3"
	"log"
)

// register provisioner factory in cluster package
func init() {
	err := cluster.RegisterProvisionerFactory("aws", "eks", &Factory{})
	if err != nil {
		log.Fatal("can't register eks provisioner")
	}
}

type Factory struct{}

func (f *Factory) NewProvisioner(cfg []byte) (cluster.Provisioner, error) {
	p := &Eks{}
	if err := yaml.Unmarshal(cfg, p); err != nil {
		return nil, fmt.Errorf("error occured during YAML unmarshalling %v", err)
	}
	if err := Validate(p); err != nil {
		return nil, err
	}

	return p, nil
}

func Validate(p *Eks) error {
	v := validator.New()
	err := v.Struct(p)

	if err != nil {
		return err.(validator.ValidationErrors)
	}

	return nil
}

type Eks struct {
	Version    string      `yaml:"version" validate:"required"`
	NodeGroups []NodeGroup `yaml:"node_group" validate:"gt=0,dive,required"`
}

type NodeGroup struct {
	Name                                string   `yaml:"name" validate:"required"`
	InstanceType                        string   `yaml:"instance_type" validate:"required"`
	OverrideInstanceTypes               []string `yaml:"override_instance_types"`
	SpotAllocationStrategy              string   `yaml:"spot_allocation_strategy"`
	SpotInstancePools                   int      `yaml:"spot_instance_pools"`
	OnDemandBaseCapacity                int      `yaml:"on_demand_base_capacity"`
	OnDemandPercentageAboveBaseCapacity int      `yaml:"on_demand_percentage_above_base_capacity"`
	AsgDesiredCapacity                  int      `yaml:"asg_desired_capacity"`
	AsgMinSize                          int      `yaml:"asg_min_size"`
	AsgMaxSize                          int      `yaml:"asg_max_size"`
	RootVolumeSize                      int      `yaml:"root_volume_size"`
	KubeletExtraArgs                    string   `yaml:"kubelet_extra_args"`
}
