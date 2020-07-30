package cluster

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

type providerFactoryMock struct {
	mock.Mock
}

func (pf *providerFactoryMock) NewProvider(cfg []byte) (Provider, error) {
	args := pf.Called(cfg)
	return args.Get(0).(Provider), args.Error(1)
}

func TestRegisterProviderFactory(t *testing.T) {
	factory := new(providerFactoryMock)

	// success on the first attempt to add into a factory
	assert.Nil(t, RegisterProviderFactory("test-type", factory))

	// error on adding a duplicate
	assert.Error(t, RegisterProviderFactory("test-type", factory))
}

func TestProviderFactoriesNewProvider(t *testing.T) {
	provider, err := NewProvider("non-existing-type", []byte{})
	assert.Nil(t, provider)
	assert.Error(t, err)

	pMock := new(providerMock)
	factory := new(providerFactoryMock)
	factory.On("NewProvider", []byte{}).Return(pMock, nil)
	assert.Nil(t, RegisterProviderFactory("registered-type", factory))
	provider, err = NewProvider("registered-type", []byte{})
	assert.Same(t, pMock, provider)
	assert.Nil(t, err)
}
