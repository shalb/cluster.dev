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
	unit := Unit{
		UnitMarkers: make(map[string]interface{}),
		Applied:     false,
		StatePtr:    &Unit{},
	}
	unit.StatePtr.UnitMarkers = unit.UnitMarkers
	err := unit.ReadConfig(spec, stack)
	if err != nil {
		log.Debug(err.Error())
		return nil, err
	}
	return &unit, nil
}

// NewFromState creates new units from state data.
func (f *Factory) NewFromState(spec map[string]interface{}, modKey string, p *project.StateProject) (project.Unit, error) {
	mod := Unit{
		UnitMarkers: make(map[string]interface{}),
		Applied:     false,
		StatePtr:    &Unit{},
	}
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
