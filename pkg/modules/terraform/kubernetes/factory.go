package kubernetes

import (
	"github.com/apex/log"
	"github.com/shalb/cluster.dev/pkg/project"
)

// Factory factory for s3 backends.
type Factory struct {
}

// New creates new module driver factory.
func (f *Factory) New(spec map[string]interface{}, infra *project.Infrastructure) (project.Module, error) {
	mod := kubernetes{
		inputs: map[string]interface{}{},
	}
	err := mod.ReadConfigCommon(spec, infra)
	if err != nil {
		log.Debug(err.Error())
		return nil, err
	}
	err = mod.ReadConfig(spec)
	if err != nil {
		log.Debug(err.Error())
		return nil, err
	}
	return &mod, nil
}

func init() {
	modDrv := Factory{}
	log.Debug("Registering module driver 'kubernetes'")
	if err := project.RegisterModuleFactory(&modDrv, "kubernetes"); err != nil {
		log.Trace("Can't register module driver 'kubernetes'.")
	}
}
