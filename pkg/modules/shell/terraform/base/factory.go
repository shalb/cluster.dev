package base

import (
	"github.com/apex/log"
	"github.com/shalb/cluster.dev/pkg/modules/shell/common"
	"github.com/shalb/cluster.dev/pkg/project"
)

// Factory factory for s3 backends.
type Factory struct {
}

func NewEmptyUnit() Unit {
	unit := Unit{
		Unit:              common.NewEmptyUnit(),
		StatePtr:          &Unit{},
		RequiredProviders: make(map[string]RequiredProvider),
		Initted:           false,
	}
	return unit
}

func NewUnit(spec map[string]interface{}, stack *project.Stack) (*Unit, error) {
	mod := NewEmptyUnit()

	cUnit, err := common.NewUnit(spec, stack)
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
	mod.BackendName = stack.BackendName
	return &mod, nil
}

// New creates new unit driver factory.
func (f *Factory) New(spec map[string]interface{}, stack *project.Stack) (project.Unit, error) {
	return NewUnit(spec, stack)
}

// NewFromState creates new unit from state data.
func (f *Factory) NewFromState(spec map[string]interface{}, modKey string, p *project.StateProject) (project.Unit, error) {
	mod := NewEmptyUnit()
	err := mod.LoadState(spec, modKey, p)
	if err != nil {
		log.Debug(err.Error())
		return nil, err
	}
	// modjs, _ := utils.JSONEncodeString(mod)
	// log.Warnf("Mod from state: %v", modjs)
	return &mod, nil
}
