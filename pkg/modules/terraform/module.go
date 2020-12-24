package terraform

import (
	"github.com/shalb/cluster.dev/pkg/project"
)

// moduleTypeKey - string representation of this module type.
const moduleTypeKey = "terraform"

// TFModule describe cluster.dev module to deploy/destroy terraform modules.
type TFModule struct {
	infraPtr        *project.Infrastructure
	projectPtr      *project.Project
	BackendPtr      project.Backend
	name            string
	Type            string
	Source          string
	Inputs          map[string]interface{}
	dependencies    []*project.Dependency
	ExpectedOutputs map[string]bool
}

// Name return module name.
func (m *TFModule) Name() string {
	return m.name
}

// InfraPtr return ptr to module infrastructure.
func (m *TFModule) InfraPtr() *project.Infrastructure {
	return m.infraPtr
}

// ProjectPtr return ptr to module project.
func (m *TFModule) ProjectPtr() *project.Project {
	return m.projectPtr
}

// InfraName return module infrastructure name.
func (m *TFModule) InfraName() string {
	return m.infraPtr.Name
}

// Backend return module backend.
func (m *TFModule) Backend() project.Backend {
	return m.infraPtr.Backend
}

// ReplaceMarkers replace all templated markers with values.
func (m *TFModule) ReplaceMarkers() error {
	for _, scanner := range m.projectPtr.MarkerScaners {
		err := project.ReplaceMarkers(m.Inputs, scanner, m)
		if err != nil {
			return err
		}
	}
	return nil
}

// Dependencies return slice of module dependencies.
func (m *TFModule) Dependencies() *[]*project.Dependency {
	return &m.dependencies
}

// Self return pointer to self.
// It should be used fo access to some internal module data or methods (witch not described in Module interface) from it terraform module driver.
func (m *TFModule) Self() interface{} {
	return m
}
