package project

import (
	"fmt"
)

// Module interface for module drivers.
type Module interface {
	Name() string
	InfraPtr() *Infrastructure
	ProjectPtr() *Project
	InfraName() string
	ReplaceMarkers() error
	Dependencies() *[]*Dependency
	Build(string) error
	Apply() error
	Plan() error
	Destroy() error
	Key() string
	ExpectedOutputs() map[string]bool
	GetState() (interface{}, error)
}

type ModuleDriver interface {
	AddTemplateFunctions(projectPtr *Project) error
	GetScanners() []MarkerScanner
}

type ModuleFactory interface {
	New(map[string]interface{}, *Infrastructure) (Module, error)
}

func RegisterModuleFactory(f ModuleFactory, modType string) error {
	if _, exists := ModuleFactoriesMap[modType]; exists {
		return fmt.Errorf("module driver with provider name '%v' already exists", modType)
	}
	ModuleFactoriesMap[modType] = f
	return nil
}

var ModuleFactoriesMap = map[string]ModuleFactory{}

// Dependency describe module dependency.
type Dependency struct {
	Module     Module
	ModuleName string
	InfraName  string
	Output     string
}

// NewModule creates and return module with needed driver.
func NewModule(spec map[string]interface{}, infra *Infrastructure) (Module, error) {
	mType, ok := spec["type"].(string)
	if !ok {
		return nil, fmt.Errorf("Incorrect module type")
	}
	modDrv, exists := ModuleFactoriesMap[mType]
	if !exists {
		return nil, fmt.Errorf("Incorrect module type '%v'", mType)
	}

	return modDrv.New(spec, infra)
}
