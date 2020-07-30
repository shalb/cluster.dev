package eks

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

	yaml := `
    version: "1.2.3"
    node_group:
        - name: group1
          instance_type: test
    `
	result, err = factory.NewProvisioner([]byte(yaml))
	assert.Implements(t, (*cluster.Provisioner)(nil), result)
	assert.Equal(t, "1.2.3", result.(*Eks).Version)
	assert.Equal(t, "group1", result.(*Eks).NodeGroups[0].Name)
	assert.Equal(t, "test", result.(*Eks).NodeGroups[0].InstanceType)
	assert.Nil(t, err)
}

func TestValidate(t *testing.T) {
	assert.Error(t, Validate(&Eks{}))
	assert.Error(t, Validate(&Eks{Version: ""}))
	assert.Error(t, Validate(&Eks{Version: "1.2.3"}))
	assert.Error(t, Validate(&Eks{Version: "1.2.3", NodeGroups: []NodeGroup{}}))
	assert.Error(t, Validate(&Eks{Version: "1.2.3", NodeGroups: []NodeGroup{{Name: ""}}}))
	assert.Error(t, Validate(&Eks{Version: "1.2.3", NodeGroups: []NodeGroup{{Name: "group1"}}}))
	assert.Error(t, Validate(&Eks{Version: "1.2.3", NodeGroups: []NodeGroup{{Name: "group1", InstanceType: ""}}}))
	assert.Error(t, Validate(&Eks{NodeGroups: []NodeGroup{{Name: "group1", InstanceType: "test"}}}))
	assert.Nil(t, Validate(&Eks{Version: "1.2.3", NodeGroups: []NodeGroup{{Name: "group1", InstanceType: "test"}}}))
}
