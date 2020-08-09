package aws

import "fmt"

// RegisterModuleFactory - add module to ModulesFactories list.
func RegisterModuleFactory(moduleName string, moduleFactory OperationFactory) error {
	if _, exists := providerOperationsFactories["modules"][moduleName]; exists {
		return fmt.Errorf("module '%s' is already registered", moduleName)
	}
	providerOperationsFactories["modules"][moduleName] = moduleFactory
	return nil
}

// GetModulesFactories return registered modules factories (used by provisioner).
func GetModulesFactories() map[string]OperationFactory {
	return providerOperationsFactories["modules"]
}
