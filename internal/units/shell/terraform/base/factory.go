package base

import (
	"github.com/apex/log"
	"github.com/shalb/cluster.dev/internal/project"
	"github.com/shalb/cluster.dev/internal/units/shell/common"
)

// Factory factory for s3 backends.
type Factory struct {
}

func NewEmptyUnit() *Unit {
	unit := Unit{
		Unit:              *common.NewEmptyUnit(),
		RequiredProviders: make(map[string]RequiredProvider),
		InitDone:          false,
	}
	return &unit
}

func NewUnit(spec map[string]interface{}, stack *project.Stack) (*Unit, error) {
	tfBase := NewEmptyUnit()

	cUnit, err := common.NewUnit(spec, stack)
	if err != nil {
		log.Debug(err.Error())
		return nil, err
	}
	tfBase.Unit = *cUnit
	err = tfBase.ReadConfig(spec, stack)
	if err != nil {
		log.Debug(err.Error())
		return nil, err
	}
	tfBase.BackendName = stack.BackendName
	return tfBase, nil
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
	return unit, nil
}
