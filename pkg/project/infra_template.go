package project

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

type InfraTemplate struct {
	Name    string                   `yaml:"name"`
	Kind    string                   `yaml:"kind"`
	Modules []map[string]interface{} `yaml:"modules"`
}

func NewInfraTemplate(data []byte) (*InfraTemplate, error) {
	iTmpl := InfraTemplate{}

	err := yaml.Unmarshal(data, &iTmpl)
	if err != nil {
		return nil, fmt.Errorf("unmarshal template data: %v", err.Error())
	}
	if len(iTmpl.Modules) < 1 {
		return nil, fmt.Errorf("parsing template: template must contain at least one module")
	}
	if iTmpl.Name == "" {
		return nil, fmt.Errorf("parsing template: template must contain 'name' field")
	}
	if iTmpl.Kind != "InfraTemplate" {
		return nil, fmt.Errorf("parsing template: unknown template object kind or kind is not set: '%v'", iTmpl.Kind)
	}
	// i.TemplateSrc = src
	return &iTmpl, nil
}
