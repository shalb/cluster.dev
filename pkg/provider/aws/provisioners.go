package aws

import (
	"fmt"
)

// RegisterProvisionerFactory - add provisioner to provisionersFactories list.
func RegisterProvisionerFactory(provisionerType string, provisionerFactory OperationFactory) error {
	if _, exists := providerOperationsFactories["provisioners"][provisionerType]; exists {
		return fmt.Errorf("provisioner '%s' is already registered", provisionerType)
	}
	providerOperationsFactories["provisioners"][provisionerType] = provisionerFactory
	return nil
}
