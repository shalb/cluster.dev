package cluster

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

type providerMock struct {
	mock.Mock
}

func (p *providerMock) Deploy() error {
	args := p.Called()
	return args.Error(0)
}

func (p *providerMock) Destroy() error {
	args := p.Called()
	return args.Error(0)
}

func TestClusterGetConfig(t *testing.T) {
	want := &Config{Installed: true, Name: "test"}
	c := Cluster{
		Config: want,
	}
	got := c.GetConfig()
	assert.Same(t, want, got)
}

func TestClusterReconcile(t *testing.T) {
	var pMock = new(providerMock)
	pMock.On("Deploy").Return(nil)
	pMock.On("Destroy").Return(nil)

	c := Cluster{
		Config:   &Config{Installed: true},
		Provider: pMock,
	}

	assert.Nil(t, c.Reconcile())
	pMock.AssertCalled(t, "Deploy")
	pMock.AssertNotCalled(t, "Destroy")

	pMock = new(providerMock)
	pMock.On("Deploy").Return(nil)
	pMock.On("Destroy").Return(nil)

	c = Cluster{
		Config:   &Config{Installed: false},
		Provider: pMock,
	}

	assert.Nil(t, c.Reconcile())
	pMock.AssertNotCalled(t, "Deploy")
	pMock.AssertCalled(t, "Destroy")
}

func TestNewFromShouldFailOnBadConfig(t *testing.T) {
	result, err := NewFrom([]byte("Hello world"))
	assert.Nil(t, result)
	assert.Error(t, err)
}

func TestNewFromShouldFailIfOnValidation(t *testing.T) {
	badYaml := `
        missing-name: test
        provider: true
	`
	result, err := NewFrom([]byte(badYaml))
	assert.Nil(t, result)
	assert.Error(t, err)

	badYaml = `
        name: test
        missing-provider: true
    `
	result, err = NewFrom([]byte(badYaml))
	assert.Nil(t, result)
	assert.Error(t, err)

	badYaml = `
        name: test
        provider: 
            missing-type: 1
    `
	result, err = NewFrom([]byte(badYaml))
	assert.Nil(t, result)
	assert.Error(t, err)
}

func TestNewFromShouldFailOnProviderFactoryError(t *testing.T) {
	providerMock := new(providerMock)
	factoryMock := new(providerFactoryMock)
	factoryMock.On("NewProvider", []byte("type: test\n")).Return(providerMock, nil)
	yaml := `
        name: test
        provider:
            type: test
    `

	assert.Nil(t, RegisterProviderFactory("test", factoryMock))

	result, err := NewFrom([]byte(yaml))
	assert.Same(t, providerMock, result.(*Cluster).Provider)
	assert.Nil(t, err)
}
