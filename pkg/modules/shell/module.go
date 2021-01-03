package shell

import (
	"reflect"

	"github.com/shalb/cluster.dev/pkg/project"
)

// Module describe cluster.dev module.
type Module struct {
	infraPtr        *project.Infrastructure
	projectPtr      *project.Project
	BackendPtr      project.Backend
	name            string
	Type            string
	Inputs          []string
	dependencies    []*project.Dependency
	expectedOutputs map[string]bool
	scriptData      []byte
	preHook         *project.Dependency
}

// Name return module name.
func (m *Module) Name() string {
	return m.name
}

// InfraPtr return ptr to module infrastructure.
func (m *Module) InfraPtr() *project.Infrastructure {
	return m.infraPtr
}

// ProjectPtr return ptr to module project.
func (m *Module) ProjectPtr() *project.Project {
	return m.projectPtr
}

// InfraName return module infrastructure name.
func (m *Module) InfraName() string {
	return m.infraPtr.Name
}

// Backend return module backend.
func (m *Module) Backend() project.Backend {
	return m.infraPtr.Backend
}

// ReplaceMarkers replace all templated markers with values.
func (m *Module) ReplaceMarkers() error {
	for i, str := range m.Inputs {
		res, err := OutputsReplacer(reflect.ValueOf(str), m)
		if err != nil {
			return err
		}
		m.Inputs[i] = res.String()
	}
	res, err := OutputsReplacer(reflect.ValueOf(string(m.scriptData)), m)
	if err != nil {
		return err
	}
	m.scriptData = []byte(res.String())
	return nil
}

// Dependencies return slice of module dependencies.
func (m *Module) Dependencies() []*project.Dependency {
	return m.dependencies
}

// ExpectedOutputs return expected outputs.
func (m *Module) ExpectedOutputs() *map[string]bool {
	return &m.expectedOutputs
}

// Self return pointer to self.
// It should be used fo access to some internal module data or methods (witch not described in Module interface) from this module driver.
func (m *Module) Self() interface{} {
	return m
}

func (m *Module) BuildDeps() error {
	for _, dep := range m.dependencies {
		err := project.BuildDep(m, dep)
		if err != nil {
			return err
		}
	}
	if m.preHook == nil {
		return nil
	}
	err := project.BuildDep(m, m.preHook)
	if err != nil {
		return err
	}
	return nil
}

func (m *Module) PreHook() *project.Dependency {
	return m.preHook
}
