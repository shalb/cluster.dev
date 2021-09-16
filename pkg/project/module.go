package project

import (
	"fmt"
)

// Module interface for module drivers.
type Module interface {
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
}

type ModuleDriver interface {
	AddTemplateFunctions(projectPtr *Project) error
	GetScanners() []MarkerScanner
}

type ModuleFactory interface {
	New(map[string]interface{}, *Stack) (Module, error)
	NewFromState(map[string]interface{}, string, *StateProject) (Module, error)
}

func RegisterModuleFactory(f ModuleFactory, modType string) error {
	if _, exists := ModuleFactoriesMap[modType]; exists {
		return fmt.Errorf("module driver with provider name '%v' already exists", modType)
	}
	ModuleFactoriesMap[modType] = f
	return nil
}

var ModuleFactoriesMap = map[string]ModuleFactory{}

// DependencyOutput describe module dependency.
type DependencyOutput struct {
	Module     Module `json:"-"`
	ModuleName string
	StackName  string
	Output     string
	OutputData interface{} `json:"-"`
}

// NewModule creates and return module with needed driver.
func NewModule(spec map[string]interface{}, stack *Stack) (Module, error) {
	mType, ok := spec["type"].(string)
	if !ok {
		return nil, fmt.Errorf("incorrect module type")
	}
	modDrv, exists := ModuleFactoriesMap[mType]
	if !exists {
		return nil, fmt.Errorf("incorrect module type '%v'", mType)
	}

	return modDrv.New(spec, stack)
}

// NewModuleFromState creates module from saved state.
func NewModuleFromState(state map[string]interface{}, stack *Stack) (Module, error) {
	mType, ok := state["type"].(string)
	if !ok {
		return nil, fmt.Errorf("Incorrect module type")
	}
	modDrv, exists := ModuleFactoriesMap[mType]
	if !exists {
		return nil, fmt.Errorf("Incorrect module type '%v'", mType)
	}

	return modDrv.New(state, stack)
}

type ModuleState interface {
	GetType() string
}
