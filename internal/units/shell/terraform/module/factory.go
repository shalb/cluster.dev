package tfmodule

import (
	"path/filepath"

	"github.com/apex/log"
	"github.com/shalb/cluster.dev/internal/project"
	"github.com/shalb/cluster.dev/internal/units/shell/common"
	"github.com/shalb/cluster.dev/internal/units/shell/terraform/base"
)

// Factory factory for s3 backends.
type Factory struct {
}

const unitKind string = "tfmodule"

func NewEmptyUnit() Unit {
	unit := Unit{
		Unit:     *base.NewEmptyUnit(),
		UnitKind: unitKind,
	}
	return unit
}

func NewUnit(spec map[string]interface{}, stack *project.Stack) (*Unit, error) {
	unit := NewEmptyUnit()
	cUnit, err := base.NewUnit(spec, stack)
	if err != nil {
		log.Debug(err.Error())
		return nil, err
	}
	unit.Unit = *cUnit
	err = unit.ReadConfig(spec, stack)
	if err != nil {
		log.Debug(err.Error())
		return nil, err
	}
	if unit.CreateFiles != nil {
		unit.CustomFiles = &common.FilesListT{}
		for _, f := range *unit.CreateFiles {
			*unit.CustomFiles = append(*unit.CustomFiles, f)
		}
	}
	unit.LocalModuleCachePath = filepath.Join(stack.ProjectPtr.CodeCacheDir, "../", "terraform", unit.Key())
	return &unit, nil
}

// New creates new unit.
func (f *Factory) New(spec map[string]interface{}, stack *project.Stack) (project.Unit, error) {
	return NewUnit(spec, stack)
}

// NewFromState creates new unit from state data.
func (f *Factory) NewFromState(spec map[string]interface{}, unitKey string, p *project.StateProject) (project.Unit, error) {
	unit := NewEmptyUnit()
	err := unit.LoadState(spec, unitKey, p)
	if err != nil {
		log.Debug(err.Error())
		return nil, err
	}
	if unit.CreateFiles != nil {
		unit.CustomFiles = &common.FilesListT{}
		for _, f := range *unit.CreateFiles {
			*unit.CustomFiles = append(*unit.CustomFiles, f)
		}
	}
	unit.LocalModuleCachePath = filepath.Join(p.CodeCacheDir, "../", "terraform", unit.Key())
	return &unit, nil
}

func init() {
	modDrv := Factory{}
	log.Debugf("Registering unit driver '%v'", unitKind)
	if err := project.RegisterUnitFactory(&modDrv, unitKind); err != nil {
		log.Trace("Can't register unit driver '" + unitKind + "'.")
	}
}
