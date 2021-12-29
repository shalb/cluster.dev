package base

import (
	"github.com/apex/log"
	"github.com/shalb/cluster.dev/pkg/units/shell/common"
	"github.com/shalb/cluster.dev/pkg/project"
)

// Factory factory for s3 backends.
type Factory struct {
}

const unitKind string = "k8s-manifest"

func NewEmptyUnit() *Unit {
	unit := Unit{
		Unit:           *common.NewEmptyUnit(),
		ManifestsFiles: &common.FilesListT{},
		UnitKind:       unitKind,
		ApplyTemplate:  true,
	}
	return &unit
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
	return mod, nil
}

// New creates new unit driver factory.
func (f *Factory) New(spec map[string]interface{}, stack *project.Stack) (project.Unit, error) {
	return NewUnit(spec, stack)
}

// NewFromState creates new unit from state data.
func (f *Factory) NewFromState(spec map[string]interface{}, modKey string, p *project.StateProject) (project.Unit, error) {
	mod := NewEmptyUnit()
	// log.Fatal("FOOO")
	err := mod.LoadState(spec, modKey, p)
	if err != nil {
		log.Debug(err.Error())
		return nil, err
	}
	// modjs, _ := utils.JSONEncodeString(mod)
	// log.Warnf("Mod from state: %v", modjs)
	return mod, nil
}

func init() {
	modDrv := Factory{}
	log.Debugf("Registering unit driver '%v'", unitKind)
	if err := project.RegisterUnitFactory(&modDrv, unitKind); err != nil {
		log.Fatalf("Can't register unit driver '%v'.", unitKind)
	}
}
