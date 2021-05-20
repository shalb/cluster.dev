package azurerm

import (
	"github.com/apex/log"
	"github.com/shalb/cluster.dev/pkg/project"
	"gopkg.in/yaml.v3"
)

// Factory factory for backends.
type Factory struct{}

// New creates the new backend.
func (f *Factory) New(config []byte, name string, p *project.Project) (project.Backend, error) {
	bk := Backend{
		name:       name,
		ProjectPtr: p,
	}
	err := yaml.Unmarshal(config, &bk.state)
	if err != nil {
		return nil, err
	}
	return &bk, nil
}

func init() {
	log.Debug("Registering backend provider azurerm..")
	if err := project.RegisterBackendFactory(&Factory{}, "azurerm"); err != nil {
		log.Trace("Can't register backend provider azurerm.")
	}
}
