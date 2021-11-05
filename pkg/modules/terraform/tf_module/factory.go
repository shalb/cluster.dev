package tfmodule

import (
	"github.com/apex/log"
	"github.com/shalb/cluster.dev/pkg/project"
)

// Factory factory for s3 backends.
type Factory struct {
}

// New creates new unit.
func (f *Factory) New(spec map[string]interface{}, stack *project.Stack) (project.Unit, error) {
	mod := Unit{}
	err := mod.ReadConfig(spec, stack)
	if err != nil {
		log.Debug(err.Error())
		return nil, err
	}
	return &mod, nil
}

// NewFromState creates new unit from state data.
func (f *Factory) NewFromState(spec map[string]interface{}, modKey string, p *project.StateProject) (project.Unit, error) {
	mod := Unit{}
	err := mod.LoadState(spec, modKey, p)
	if err != nil {
		log.Debug(err.Error())
		return nil, err
	}
	return &mod, nil
}

// func init() {
// 	modDrv := Factory{}
// 	log.Debug("Registering unit driver 'terraform'")
// 	if err := project.RegisterUnitFactory(&modDrv, "terraform"); err != nil {
// 		log.Trace("Can't register unit driver 'terraform'.")
// 	}
// }
