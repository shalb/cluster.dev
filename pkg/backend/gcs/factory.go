package gcs

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
	err := yaml.Unmarshal(config, &bk)
	if err != nil {
		return nil, err
	}
	if bk.Prefix != "" {
		bk.Prefix += "/"
	}
	bk.state, err = getStateMap(bk)
	return &bk, nil
}

func init() {
	log.Debug("Registering backend provider gcs..")
	if err := project.RegisterBackendFactory(&Factory{}, "gcs"); err != nil {
		log.Trace("Can't register backend provider gcs.")
	}
}
