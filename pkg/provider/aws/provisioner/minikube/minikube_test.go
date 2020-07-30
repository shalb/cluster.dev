package minikube

import (
	"github.com/shalb/cluster.dev/pkg/cluster"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFactoryNewProvisioner(t *testing.T) {
	factory := new(Factory)

	result, err := factory.NewProvisioner([]byte("Hello world"))
	assert.Nil(t, result)
	assert.Error(t, err)

	result, err = factory.NewProvisioner([]byte("random-yaml: random"))
	assert.Nil(t, result)
	assert.Error(t, err)

	result, err = factory.NewProvisioner([]byte("instance_type: test"))
	assert.Implements(t, (*cluster.Provisioner)(nil), result)
	assert.Equal(t, "test", result.(*Minikube).InstanceType)
	assert.Nil(t, err)
}

func TestValidate(t *testing.T) {
	assert.Error(t, Validate(&Minikube{}))
	assert.Error(t, Validate(&Minikube{InstanceType: ""}))
	assert.Nil(t, Validate(&Minikube{InstanceType: "test"}))
}
