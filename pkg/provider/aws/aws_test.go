package aws

import (
	"github.com/shalb/cluster.dev/pkg/cluster"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

type provisionerFactoryMock struct {
	mock.Mock
}

func (pf *provisionerFactoryMock) NewProvisioner(cfg []byte) (cluster.Provisioner, error) {
	args := pf.Called(cfg)
	return args.Get(0), args.Error(1)
}

func TestFactoryNewProvider(t *testing.T) {
	provisionerFactoryMock := new(provisionerFactoryMock)
	assert.Nil(t, cluster.RegisterProvisionerFactory("aws", "test", provisionerFactoryMock))
	factory := new(Factory)

	result, err := factory.NewProvider([]byte("Hello world"))
	assert.Nil(t, result)
	assert.Error(t, err)

	result, err = factory.NewProvider([]byte("random-yaml: random"))
	assert.Nil(t, result)
	assert.Error(t, err)

	// test default values
	yaml := `
    type: test
    region: test
    provisioner:
        type: test
    `
	provisionerFactoryMock.On("NewProvisioner", []byte("type: test\n")).Return(nil, nil)
	result, err = factory.NewProvider([]byte(yaml))
	assert.Implements(t, (*cluster.Provider)(nil), result)
	assert.Equal(t, "test", result.(*Aws).Type)
	assert.Equal(t, "test", result.(*Aws).Region)
	assert.Equal(t, "cluster.dev", result.(*Aws).Domain)
	assert.Equal(t, "default", result.(*Aws).Vpc)
	assert.Nil(t, err)

	// test redefine values
	yaml = `
    type: test
    region: region-test
    vpc: vpc-test
    domain: domain-test
    availability_zones:
        - test1
        - test2
    provisioner:
        type: test
    `
	provisionerFactoryMock.On("NewProvisioner", []byte("type: test\n")).Return(nil, nil)
	result, err = factory.NewProvider([]byte(yaml))
	assert.Implements(t, (*cluster.Provider)(nil), result)
	assert.Equal(t, "test", result.(*Aws).Type)
	assert.Equal(t, "region-test", result.(*Aws).Region)
	assert.Equal(t, "domain-test", result.(*Aws).Domain)
	assert.Equal(t, "vpc-test", result.(*Aws).Vpc)
	assert.Equal(t, []string{"test1", "test2"}, result.(*Aws).AvailabilityZones)
	assert.Nil(t, err)

}

func TestValidate(t *testing.T) {
	assert.Error(t, Validate(&Aws{}))
	assert.Error(t, Validate(&Aws{Region: ""}))
	assert.Error(t, Validate(&Aws{Type: ""}))
	assert.Error(t, Validate(&Aws{Region: "", Type: "test"}))
	assert.Error(t, Validate(&Aws{Region: "test", Type: ""}))
	assert.Error(t, Validate(&Aws{Region: "test", Type: "test"}))
	assert.Nil(t, Validate(&Aws{Region: "test", Type: "test", Provisioner: struct{}{}}))
}
