package common

import (
	"github.com/apex/log"
	"github.com/shalb/cluster.dev/pkg/project"
)

// Factory factory for s3 backends.
type Factory struct {
}

const unitKind string = "shell"

// NewEmptyUnit creates new unit.
func NewEmptyUnit() *Unit {
	unit := Unit{
		//UnitMarkers: make(map[string]interface{}),
		AlreadyApplied:   false,
		UnitKind:         unitKind,
		CreateFiles:      &FilesListT{},
		DependenciesList: project.NewUnitLinksT(),
		FApply:           false,
		Env:              make(map[string]string),
	}
	unit.OutputParsers = map[string]OutputParser{
		"json":      unit.JSONOutputParser,
		"regexp":    unit.RegexOutputParser,
		"separator": unit.SeparatorOutputParser,
	}
	//unit.StatePtr.UnitMarkers = unit.UnitMarkers
	return &unit
}

// NewUnit creates new unit and load config.
func NewUnit(spec map[string]interface{}, stack *project.Stack) (*Unit, error) {
	unit := NewEmptyUnit()
	//unit.StatePtr.UnitMarkers = unit.UnitMarkers
	unit.BackendName = stack.Backend.Name()
	err := unit.ReadConfig(spec, stack)
	if err != nil {
		log.Debug(err.Error())
		return nil, err
	}
	return unit, nil
}

// New creates new units driver factory.
func (f *Factory) New(spec map[string]interface{}, stack *project.Stack) (project.Unit, error) {
	return NewUnit(spec, stack)
}

// NewFromState creates new units from state data.
func (f *Factory) NewFromState(spec map[string]interface{}, modKey string, p *project.StateProject) (project.Unit, error) {
	mod := NewEmptyUnit()
	err := mod.LoadState(spec, modKey, p)
	if err != nil {
		log.Debug(err.Error())
		return nil, err
	}

	return mod, nil
}

func init() {
	modDrv := Factory{}
	log.Debugf("Registering unit driver '%v'", unitKind)
	if err := project.RegisterUnitFactory(&modDrv, unitKind); err != nil {
		log.Trace("Can't register unit driver '" + unitKind + "'.")
	}
}
