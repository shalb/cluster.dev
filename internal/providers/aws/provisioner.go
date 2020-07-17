package aws

import (
	"fmt"
	"time"
)

// ProvisionerCommon - interface for all provisioners.
type ProvisionerCommon interface {
	Deploy(time.Duration) error
	Destroy() error
	GetKubeConfig() (string, error)
	PullKubeConfig() error
}

// NewProvisioner create new provisioner instance.
func NewProvisioner(conf providerConfSpec) (ProvisionerCommon, error) {

	provisionerType, ok := conf.Provisioner["type"].(string)
	if !ok {
		return nil, fmt.Errorf("can't determinate provisioner type")
	}
	var pv ProvisionerCommon
	var err error
	switch provisionerType {
	case "minikube":
		pv, err = NewProvisionerMinikube(conf)
		return pv, err
	case "eks":
		pv, err = NewProvisionerEks(conf)
		return pv, err
	default:
		return nil, fmt.Errorf("unknown provisioner type '%s'", provisionerType)
	}
}
