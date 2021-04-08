package project

import (
	"fmt"

	"github.com/Masterminds/semver"
	"github.com/apex/log"
	"github.com/shalb/cluster.dev/pkg/config"
	"gopkg.in/yaml.v3"
)

const infraTemplateObjKindKey = "InfraTemplate"

type InfraTemplate struct {
	Name             string                   `yaml:"name"`
	Kind             string                   `yaml:"kind"`
	Modules          []map[string]interface{} `yaml:"modules"`
	ReqClientVersion string                   `yaml:"cliVersion"`
}

func NewInfraTemplate(data []byte) (*InfraTemplate, error) {
	iTmpl := InfraTemplate{}
	err := yaml.Unmarshal(data, &iTmpl)
	if err != nil {
		if config.Global.TraceLog {
			log.Debug(string(data))
		}
		return nil, fmt.Errorf("unmarshal template data: %v", err.Error())
	}
	if len(iTmpl.Modules) < 1 {
		return nil, fmt.Errorf("parsing template: template must contain at least one module")
	}
	if iTmpl.Name == "" {
		return nil, fmt.Errorf("parsing template: template must contain 'name' field")
	}
	if iTmpl.Kind != infraTemplateObjKindKey {
		return nil, fmt.Errorf("parsing template: unknown template object kind or kind is not set: '%v'", iTmpl.Kind)
	}
	log.Debug("check client version")
	if iTmpl.ReqClientVersion != "" {
		reqVerConstraints, err := semver.NewConstraint(iTmpl.ReqClientVersion)
		if err != nil {
			return nil, fmt.Errorf("parsing template: can't parse required client version: %v", iTmpl.ReqClientVersion)
		}
		ver, err := semver.NewVersion(config.Global.Version)
		if err != nil {
			// Invalid curent cli version. May be test revision.
			// TODO need check!!
			return nil, fmt.Errorf("parsing template: internalcan't parse client version: %v", iTmpl.ReqClientVersion)
		}
		ok, messages := reqVerConstraints.Validate(ver)
		log.Debugf("Validating version: cli: %v, req: %v", ver, iTmpl.ReqClientVersion)
		if !ok {
			errReasons := ""
			for _, msg := range messages {
				errReasons += fmt.Sprintf("%v\n", msg)
			}
			return nil, fmt.Errorf("cdev template version validatin error: \n%v", errReasons)
		}
	}
	// i.TemplateSrc = src
	return &iTmpl, nil
}
