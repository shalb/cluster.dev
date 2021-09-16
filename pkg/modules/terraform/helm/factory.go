package helm

import (
	"github.com/apex/log"
	"github.com/shalb/cluster.dev/pkg/project"
)

// Factory factory for s3 backends.
type Factory struct {
}

// New creates new module driver factory.
func (f *Factory) New(spec map[string]interface{}, stack *project.Stack) (project.Module, error) {
	mod := Module{
		helmOpts: map[string]interface{}{},
		sets:     map[string]interface{}{},
	}
	err := mod.ReadConfig(spec, stack)
	if err != nil {
		log.Debug(err.Error())
		return nil, err
	}
	return &mod, nil
}

// NewFromState creates new module from state data.
func (f *Factory) NewFromState(spec map[string]interface{}, modKey string, p *project.StateProject) (project.Module, error) {
	mod := Module{}
	err := mod.LoadState(spec, modKey, p)
	if err != nil {
		log.Debug(err.Error())
		return nil, err
	}
	return &mod, nil
}

func init() {
	modDrv := Factory{}
	log.Debug("Registering module driver 'helm'")
	if err := project.RegisterModuleFactory(&modDrv, "helm"); err != nil {
		log.Trace("Can't register module driver 'helm'.")
	}
}
