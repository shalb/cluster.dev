package cluster

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"gopkg.in/yaml.v3"
)

type Reconciler interface {
	GetConfig() *Config
	Reconcile() error
}

type Config struct {
	Name      string          `json:"name" validate:"required,ascii"`
	Installed bool            `json:"installed"`
	Provider  interface{}     `json:"provider" validate:"required"`
	Addons    map[string]bool `json:"addons"`
	Apps      []string        `json:"apps"`
}

type Cluster struct {
	Config   *Config
	Provider Provider
}

func (c *Cluster) GetConfig() *Config {
	return c.Config
}

func (c *Cluster) Reconcile() error {
	if c.Config.Installed {
		if err := c.Provider.Deploy(); err != nil {
			return err
		}
	} else {
		if err := c.Provider.Destroy(); err != nil {
			return err
		}
	}
	return nil
}

func Validate(cfg *Config) error {
	v := validator.New()
	err := v.Struct(cfg)

	if err != nil {
		return err.(validator.ValidationErrors)
	}
	return nil
}

//var newProvider = func (providerType string, provYaml []byte) (*Provider, error)

func NewFrom(in []byte) (Reconciler, error) {
	cfg := &Config{}
	if err := yaml.Unmarshal(in, cfg); err != nil {
		return nil, fmt.Errorf("error occured during YAML unmarshalling %v", err)
	}

	if err := Validate(cfg); err != nil {
		return nil, err
	}

	providerType, exists := cfg.Provider.(map[string]interface{})["type"].(string)
	if !exists {
		return nil, fmt.Errorf("YAML must contain provider.type field")
	}

	provYaml, err := yaml.Marshal(cfg.Provider)
	if err != nil {
		return nil, err
	}

	p, err := NewProvider(providerType, provYaml)
	if err != nil {
		return nil, err
	}

	return &Cluster{
		Config:   cfg,
		Provider: p,
	}, nil
}
