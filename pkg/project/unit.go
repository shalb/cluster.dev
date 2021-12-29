package project

import (
	"fmt"
	"sync"

	"github.com/apex/log"
)

// Unit interface for unit drivers.
type Unit interface {
	Name() string
	Stack() *Stack
	Project() *Project
	Backend() Backend
	Prepare() error // Prepare scan all markers in unit, and build project unit links, and unit dependencies.
	Dependencies() *UnitLinksT
	Build() error
	Init() error
	Apply() error
	Plan() error
	Destroy() error
	Key() string
	GetState() interface{}
	GetDiffData() interface{}
	GetStateDiffData() interface{}
	LoadState(interface{}, string, *StateProject) error
	KindKey() string
	CodeDir() string
	UpdateProjectRuntimeData(p *Project) error
	WasApplied() bool
	ForceApply() bool
	Mux() *sync.Mutex
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

// NewUnit creates and return unit with needed driver.
func NewUnit(spec map[string]interface{}, stack *Stack) (Unit, error) {
	mType, ok := spec["type"].(string)
	if !ok {
		return nil, fmt.Errorf("incorrect unit type")
	}
	uName, ok := spec["name"].(string)
	modDrv, exists := UnitFactoriesMap[mType]
	// TODO remove deprecated unit type 'kubernetes'
	if mType == "kubernetes" {
		log.Warnf("Unit: '%v'. Unit type 'kubernetes' is deprecated and will be removed in future releases. Use 'k8s-manifest' instead", fmt.Sprintf("%v.%v", stack.Name, uName))
	}
	if !exists {
		// TODO remove deprecated unit type 'terraform'
		if mType == "terraform" {
			log.Warnf("Unit: '%v'. Unit type 'terraform' is deprecated and will be removed in future releases. Use 'tfmodule' instead", fmt.Sprintf("%v.%v", stack.Name, uName))
			modDrv = UnitFactoriesMap["tfmodule"]
		} else {
			return nil, fmt.Errorf("new unit: bad unit type in state '%v'", mType)
		}
	}

	return modDrv.New(spec, stack)
}

// NewUnitFromState creates unit from saved state.
func NewUnitFromState(state map[string]interface{}, name string, p *StateProject) (Unit, error) {
	mType, ok := state["type"].(string)
	if !ok {
		return nil, fmt.Errorf("internal error: unit type field in state does not found")
	}
	modDrv, exists := UnitFactoriesMap[mType]
	if !exists {
		// TODO remove deprecated unit type 'terraform'
		if mType == "terraform" {
			modDrv = UnitFactoriesMap["tfmodule"]
		} else {
			return nil, fmt.Errorf("internal error: bad unit type in state '%v'", mType)
		}
	}
	return modDrv.NewFromState(state, name, p)
}

type UnitState interface {
	GetType() string
}
