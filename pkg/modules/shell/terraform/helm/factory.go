package helm

import (
	"github.com/apex/log"
	"github.com/shalb/cluster.dev/pkg/modules/shell/terraform/base"
	"github.com/shalb/cluster.dev/pkg/project"
)

// Factory factory for s3 backends.
type Factory struct {
}

const unitKind string = "helm"

func NewEmptyUnit() Unit {
	unit := Unit{
		Unit:     *base.NewEmptyUnit(),
		HelmOpts: map[string]interface{}{},
		Sets:     map[string]interface{}{},
		UnitKind: unitKind,
	}
	return unit
}

func NewUnit(spec map[string]interface{}, stack *project.Stack) (*Unit, error) {
	mod := NewEmptyUnit()
	cUnit, err := base.NewUnit(spec, stack)
	if err != nil {
		log.Debug(err.Error())
		return nil, err
	}
	mod.Unit = *cUnit
	err = mod.ReadConfig(spec, stack)
	if err != nil {
		log.Debug(err.Error())
		return nil, err
	}
	return &mod, nil
}

// New creates new unit driver factory.
func (f *Factory) New(spec map[string]interface{}, stack *project.Stack) (project.Unit, error) {
	return NewUnit(spec, stack)
}

// NewFromState creates new unit from state data.
func (f *Factory) NewFromState(spec map[string]interface{}, modKey string, p *project.StateProject) (project.Unit, error) {
	unit := NewEmptyUnit()
	err := unit.LoadState(spec, modKey, p)
	if err != nil {
		log.Debug(err.Error())
		return nil, err
	}
	return &unit, nil
}

func init() {
	modDrv := Factory{}
	log.Debugf("Registering unit driver '%v'", unitKind)
	if err := project.RegisterUnitFactory(&modDrv, unitKind); err != nil {
		log.Trace("Can't register unit driver '" + unitKind + "'.")
	}
}
