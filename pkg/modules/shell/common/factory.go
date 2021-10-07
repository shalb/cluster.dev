package common

import (
	"github.com/apex/log"
	"github.com/shalb/cluster.dev/pkg/project"
)

// Factory factory for s3 backends.
type Factory struct {
}

// New creates new units driver factory.
func (f *Factory) New(spec map[string]interface{}, stack *project.Stack) (project.Unit, error) {
	mod := Unit{
		markers: make(map[string]interface{}),
		applied: false,
	}
	mod.outputParsers = map[string]outputParser{
		"json":      mod.JSONOutputParser,
		"regexp":    mod.RegexOutputParser,
		"separator": mod.SeparatorOutputParser,
	}
	err := mod.ReadConfig(spec, stack)
	if err != nil {
		log.Debug(err.Error())
		return nil, err
	}
	return &mod, nil
}

// NewFromState creates new units from state data.
func (f *Factory) NewFromState(spec map[string]interface{}, modKey string, p *project.StateProject) (project.Unit, error) {
	mod := Unit{}
	err := mod.LoadState(spec, modKey, p)
	if err != nil {
		log.Debug(err.Error())
		return nil, err
	}
	return &mod, nil
}

func init() {
	modDrv := Factory{}
	if err := project.RegisterUnitFactory(&modDrv, "shell"); err != nil {
		log.Trace("Can't register unit driver 'shell'.")
	}
}
