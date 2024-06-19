package project

import (
	"fmt"

	"github.com/Masterminds/semver"
	"github.com/apex/log"
	"github.com/shalb/cluster.dev/internal/config"
	"github.com/shalb/cluster.dev/pkg/utils"
	"gopkg.in/yaml.v3"
)

const stackTemplateObjKindKey = "StackTemplate"

type stackTemplate struct {
	Name             string                   `yaml:"name"`
	Kind             string                   `yaml:"kind"`
	Units            []map[string]interface{} `yaml:"units"`
	Modules          []map[string]interface{} `yaml:"modules,omitempty"`
	ReqClientVersion string                   `yaml:"cliVersion"`
}

func NewStackTemplate(data []byte) (*stackTemplate, error) {
	iTmpl := stackTemplate{}
	err := yaml.Unmarshal(data, &iTmpl)
	if err != nil {
		return nil, fmt.Errorf("unmarshal template data: %v", utils.ResolveYamlError(data, err))
	}
	if len(iTmpl.Units) < 1 {
		if len(iTmpl.Modules) < 1 {
			return nil, fmt.Errorf("parsing template: template must contain at least one unit")
		}
		iTmpl.Units = iTmpl.Modules
		iTmpl.Modules = nil
		log.Warnf("'modules' key is deprecated and will be removed in future releases. Use 'units' instead")
	}
	if iTmpl.Name == "" {
		return nil, fmt.Errorf("parsing template: template must contain 'name' field")
	}
	if iTmpl.Kind != stackTemplateObjKindKey {
		if iTmpl.Kind != "InfraTemplate" {
			return nil, fmt.Errorf("parsing template: unknown template object kind or kind is not set: '%v'", iTmpl.Kind)
		}
		log.Warnf("'InfraTemplate' kind is deprecated and will be removed in future releases. Use 'StackTemplate' instead")
	}
	if iTmpl.ReqClientVersion != "" {
		log.Debug("Checking client version...")
		reqVerConstraints, err := semver.NewConstraint(iTmpl.ReqClientVersion)
		if err != nil {
			return nil, fmt.Errorf("parsing template: can't parse required client version: %v, err: %v", iTmpl.ReqClientVersion, err)
		}
		ver, err := semver.NewVersion(config.Global.Version)
		if err != nil {
			// Invalid current cli version. Maybe test revision.
			return nil, fmt.Errorf("parsing template: can't parse client version: %v", err)
		}
		ok, messages := reqVerConstraints.Validate(ver)
		errReasons := ""
		if !ok {
			if len(messages) > 1 {
				for _, msg := range messages {
					errReasons += fmt.Sprintf("cdev version: %v\n", msg)
				}
			} else {
				errReasons = messages[0].Error()
			}
			return nil, fmt.Errorf("cdev template version validation error: template: '%v', cli: '%v', req: '%v', message: '%v'", iTmpl.Name, ver, iTmpl.ReqClientVersion, errReasons)
		}
		log.Debugf("Version validated: stack: %v cli: %v, req: %v", iTmpl.Name, ver, iTmpl.ReqClientVersion)
	}
	// i.TemplateSrc = src
	return &iTmpl, nil
}
