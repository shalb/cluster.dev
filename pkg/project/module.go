package project

import (
	"fmt"
)

// Module interface for module drivers.
type Module interface {
	CreateCodeDir(string) error
	Name() string
	InfraPtr() *Infrastructure
	ProjectPtr() *Project
	InfraName() string
	ReplaceMarkers() error
	GetApplyShellCmd() string
	GetDestroyShellCmd() string
	Dependencies() *[]*Dependency
	Self() interface{}
}

type ModuleDriver interface {
	GetTemplateFunctions() map[string]interface{}
	GetScanners() []MarkerScanner
	NewModule(map[string]interface{}, *Infrastructure) (Module, error)
}

// ModuleDriverFactory - interface for module driver factory. New() creates module driver.
type ModuleDriverFactory interface {
	New(*Project) ModuleDriver
}

// RegisterModuleDriverFactory - register factory of some driver type (like terraform) in map.
func RegisterModuleDriverFactory(modDrv ModuleDriverFactory, driverType string) error {
	if _, exists := ModuleDriverFactories[driverType]; exists {
		return fmt.Errorf("module driver with provider name '%v' already exists", driverType)
	}
	ModuleDriverFactories[driverType] = modDrv
	return nil
}

// ModuleDriverFactories map of module drivers factories. Use ModulesFactories["type"].New() to create module driver of type 'type'
var ModuleDriverFactories = map[string]ModuleDriverFactory{}

// Dependency describe module dependency.
type Dependency struct {
	Module     *Module
	ModuleName string
	InfraName  string
	Output     string
}

// NewModule creates and return module with needed driver.
func NewModule(spec map[string]interface{}, infra *Infrastructure) (Module, error) {
	mType, ok := spec["type"]
	if !ok {
		return nil, fmt.Errorf("Incorrect module type")
	}
	modDrv, exists := infra.ProjectPtr.ModuleDrivers[mType.(string)]
	if !exists {
		return nil, fmt.Errorf("Incorrect module type '%v'", mType)
	}

	return modDrv.NewModule(spec, infra)
}
