package gcs

import (
	"github.com/apex/log"
	"github.com/shalb/cluster.dev/pkg/project"
	"github.com/shalb/cluster.dev/pkg/utils"
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
	state := map[string]interface{}{}
	err := yaml.Unmarshal(config, &bk)
	if err != nil {
		return nil, utils.ResolveYamlError(config, err)
	}
	if bk.Prefix != "" {
		bk.Prefix += "/"
	}
	bk.state = state
	return &bk, bk.Configure()
}

func init() {
	log.Debug("Registering backend provider gcs..")
	if err := project.RegisterBackendFactory(&Factory{}, "gcs"); err != nil {
		log.Trace("Can't register backend provider gcs.")
	}
}
