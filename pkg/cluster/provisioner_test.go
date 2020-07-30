package cluster

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

type provisionerFactoryMock struct {
	mock.Mock
}

func (pf *provisionerFactoryMock) NewProvisioner(cfg []byte) (Provisioner, error) {
	args := pf.Called(cfg)
	return args.Get(0).(Provisioner), args.Error(1)
}

func TestRegisterProvisionerFactory(t *testing.T) {
	factory := new(provisionerFactoryMock)

	// success on the first attempt to add into a factory
	assert.Nil(t, RegisterProvisionerFactory("test-type", "test-type", factory))

	// error on adding a duplicate
	assert.Error(t, RegisterProvisionerFactory("test-type", "test-type", factory))
}

func TestProvisionerFactoriesNewProvisioner(t *testing.T) {
	provisioner, err := NewProvisioner("non-existing-type", "test-type", []byte{})
	assert.Nil(t, provisioner)
	assert.Error(t, err)

	provisioner, err = NewProvisioner("test-type", "non-existing-type", []byte{})
	assert.Nil(t, provisioner)
	assert.Error(t, err)

	pMock := new(providerMock)
	factory := new(provisionerFactoryMock)
	factory.On("NewProvisioner", []byte{}).Return(pMock, nil)
	assert.Nil(t, RegisterProvisionerFactory("registered-type", "registered-type", factory))
	provisioner, err = NewProvisioner("registered-type", "registered-type", []byte{})
	assert.Same(t, pMock, provisioner)
	assert.Nil(t, err)
}
