package s3

import (
	"github.com/apex/log"
	"github.com/shalb/cluster.dev/pkg/project"
	"gopkg.in/yaml.v3"
)

// Factory factory for s3 backends.
type Factory struct{}

// New creates the new s3 backend.
func (f *Factory) New(config []byte, name string, p *project.Project) (project.Backend, error) {
	bk := Backend{
		name:       name,
		ProjectPtr: p,
	}
	state := map[string]interface{}{}
	err := yaml.Unmarshal(config, &bk)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(config, &state)
	if err != nil {
		return nil, err
	}
	bk.state = state
	return &bk, nil
}

func init() {
	log.Debug("Registering backend provider s3..")
	if err := project.RegisterBackendFactory(&Factory{}, "s3"); err != nil {
		log.Trace("Can't register backend provider s3.")
	}
}
