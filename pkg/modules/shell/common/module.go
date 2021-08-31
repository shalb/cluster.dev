package common

import (
	"fmt"

	"github.com/shalb/cluster.dev/pkg/project"
)

// Module describe cluster.dev module to deploy/destroy terraform modules.
type Module struct {
	infraPtr        *project.Infrastructure
	projectPtr      *project.Project
	backendPtr      project.Backend
	name            string
	dependencies    []*project.Dependency
	expectedOutputs map[string]bool
	codeDir         string
	filesList       map[string][]byte
	specRaw         map[string]interface{}
	markers         map[string]interface{}
	applyOutput     []byte
}

func (m *Module) AddRequiredProvider(name, source, version string) {
}

func (m *Module) Markers() map[string]interface{} {
	return m.markers
}

func (m *Module) FilesList() map[string][]byte {
	return m.filesList
}

func (m *Module) ReadConfig(spec map[string]interface{}, infra *project.Infrastructure) error {
	return nil
}

func (m *Module) ExpectedOutputs() map[string]bool {
	return m.expectedOutputs
}

// Name return module name.
func (m *Module) Name() string {
	return m.name
}

// InfraPtr return ptr to module infrastructure.
func (m *Module) InfraPtr() *project.Infrastructure {
	return m.infraPtr
}

// ApplyOutput return output of last module applying.
func (m *Module) ApplyOutput() []byte {
	return m.applyOutput
}

// Outputs module.
func (m *Module) Outputs() (string, error) {
	return "", nil
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

// Dependencies return slice of module dependencies.
func (m *Module) Dependencies() *[]*project.Dependency {
	return &m.dependencies
}

func (m *Module) Init() error {

	return nil
}

func (m *Module) Apply() error {
	return nil
}

// Plan module.
func (m *Module) Plan() error {
	return nil
}

// Destroy module.
func (m *Module) Destroy() error {
	return nil
}

// Key return uniq module index (string key for maps).
func (m *Module) Key() string {
	return fmt.Sprintf("%v.%v", m.InfraName(), m.name)
}

// CodeDir return path to module code directory.
func (m *Module) CodeDir() string {
	return m.codeDir
}

// UpdateProjectRuntimeData update project runtime dataset, adds module outputs.
// TODO: get module outputs and write to project runtime dataset. Now this function is only for printer's module interface.
func (m *Module) UpdateProjectRuntimeData(p *project.Project) error {
	return nil
}

// ReplaceMarkers replace all templated markers with values.
func (m *Module) ReplaceMarkers() error {
	return nil
}

func (m *Module) KindKey() string {
	return "shell"
}
