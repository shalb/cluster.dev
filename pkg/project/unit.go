package project

import (
	"fmt"
)

// Unit interface for unit drivers.
type Unit interface {
	Name() string
	StackPtr() *Stack
	ProjectPtr() *Project
	StackName() string
	ReplaceMarkers() error
	Dependencies() *[]*DependencyOutput
	Build() error
	Apply() error
	Plan() error
	Destroy() error
	Key() string
	ExpectedOutputs() map[string]*DependencyOutput
	GetState() interface{}
	GetDiffData() interface{}
	LoadState(interface{}, string, *StateProject) error
	KindKey() string
	CodeDir() string
	Markers() map[string]interface{}
	UpdateProjectRuntimeData(p *Project) error
	WasApplied() bool
}

type UnitDriver interface {
	AddTemplateFunctions(projectPtr *Project) error
	GetScanners() []MarkerScanner
}

type UnitFactory interface {
	New(map[string]interface{}, *Stack) (Unit, error)
	NewFromState(map[string]interface{}, string, *StateProject) (Unit, error)
}

func RegisterUnitFactory(f UnitFactory, modType string) error {
	if _, exists := UnitFactoriesMap[modType]; exists {
		return fmt.Errorf("unit driver with provider name '%v' already exists", modType)
	}
	UnitFactoriesMap[modType] = f
	return nil
}

var UnitFactoriesMap = map[string]UnitFactory{}

// DependencyOutput describe unit dependency.
type DependencyOutput struct {
	Unit       Unit `json:"-"`
	UnitName   string
	StackName  string
	Output     string
	OutputData interface{}
}

// NewUnit creates and return unit with needed driver.
func NewUnit(spec map[string]interface{}, stack *Stack) (Unit, error) {
	mType, ok := spec["type"].(string)
	if !ok {
		return nil, fmt.Errorf("incorrect unit type")
	}
	modDrv, exists := UnitFactoriesMap[mType]
	if !exists {
		return nil, fmt.Errorf("incorrect unit type '%v'", mType)
	}

	return modDrv.New(spec, stack)
}

// NewUnitFromState creates unit from saved state.
func NewUnitFromState(state map[string]interface{}, stack *Stack) (Unit, error) {
	mType, ok := state["type"].(string)
	if !ok {
		return nil, fmt.Errorf("Incorrect unit type")
	}
	modDrv, exists := UnitFactoriesMap[mType]
	if !exists {
		return nil, fmt.Errorf("Incorrect unit type '%v'", mType)
	}

	return modDrv.New(state, stack)
}

type UnitState interface {
	GetType() string
}
