package do

import (
	"fmt"

	"github.com/apex/log"
	"github.com/shalb/cluster.dev/pkg/project"
	"github.com/shalb/cluster.dev/pkg/utils"
	"gopkg.in/yaml.v3"
)

// Factory factory for do backends.
type Factory struct{}

// New creates the new do backend.
func (f *Factory) New(config []byte, name string, p *project.Project) (project.Backend, error) {
	bk := Backend{
		name:       name,
		ProjectPtr: p,
	}
	state := map[string]interface{}{}
	err := yaml.Unmarshal(config, &bk)
	if err != nil {
		return nil, utils.ResolveYamlError(config, err)
	}
	err = yaml.Unmarshal(config, &state)
	if err != nil {
		return nil, utils.ResolveYamlError(config, err)
	}
	bk.state = state
	if bk.AccessKey != "" && bk.SecretKey != "" {
		return &bk, nil
	}
	if bk.AccessKey == "" && bk.SecretKey == "" {
		return &bk, nil
	}
	return nil, fmt.Errorf("error in backend '%v' configuration", bk.Name())
}

func init() {
	log.Debug("Registering backend provider do..")
	if err := project.RegisterBackendFactory(&Factory{}, "do"); err != nil {
		log.Fatalf("Can't register backend provider do. %v", err.Error())
	}
}
