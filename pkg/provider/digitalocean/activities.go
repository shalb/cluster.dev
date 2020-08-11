package digitalocean

import (
	"fmt"

	"github.com/shalb/cluster.dev/pkg/cluster"
	"github.com/shalb/cluster.dev/pkg/provider"
)

// Base of provider operations (modules and provisioners).
var providerActivitiesFactories = map[string]map[string]ActivityFactory{
	"modules":      make(map[string]ActivityFactory),
	"provisioners": make(map[string]ActivityFactory),
}

// RegisterActivityFactory - add module to ModulesFactories list.
func RegisterActivityFactory(aType, aName string, moduleFactory ActivityFactory) error {
	if _, exists := providerActivitiesFactories[aType][aName]; exists {
		return fmt.Errorf("activity '%s.%s' is already registered", aType, aName)
	}

	providerActivitiesFactories[aType][aName] = moduleFactory
	return nil
}

// ActivityFactory common interface for modules and provisioners factories.
type ActivityFactory interface {
	New(Config, *cluster.State) (provider.Activity, error)
}

// GetModulesFactories return registered modules factories (used by provisioner).
func GetModulesFactories() map[string]ActivityFactory {
	return providerActivitiesFactories["modules"]
}
