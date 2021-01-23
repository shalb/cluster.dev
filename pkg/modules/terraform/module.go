package terraform

import (
	"fmt"

	"github.com/shalb/cluster.dev/internal/executor"
	"github.com/shalb/cluster.dev/pkg/project"
)

// moduleTypeKey - string representation of this module type.
const moduleTypeKey = "terraform"

// TFModule describe cluster.dev module to deploy/destroy terraform modules.
type TFModule struct {
	infraPtr                *project.Infrastructure
	projectPtr              *project.Project
	BackendPtr              project.Backend
	name                    string
	Type                    string
	Source                  string
	Inputs                  map[string]interface{}
	dependenciesRemoteState []*project.Dependency
	expectedRemoteStates    map[string]bool
	preHook                 []byte
	postHook                []byte
	codeDir                 string
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
	err := project.ScanMarkers(m.Inputs, yamlBlockMarkerScanner, m)
	if err != nil {
		return err
	}
	err = project.ScanMarkers(m.Inputs, remoteStatesScanner, m)
	if err != nil {
		return err
	}
	return nil
}

// Dependencies return slice of module dependencies.
func (m *TFModule) Dependencies() []*project.Dependency {
	depsUniq := map[*project.Dependency]bool{}
	for _, dep := range m.dependenciesRemoteState {
		depsUniq[dep] = true
	}
	deps := []*project.Dependency{}
	for dep := range depsUniq {
		deps = append(deps, dep)
	}
	return deps
}

// Self return pointer to self.
// It should be used fo access to some internal module data or methods (witch not described in Module interface) from it terraform module driver.
func (m *TFModule) Self() interface{} {
	return m
}

// BuildDeps check all dependencies and add module pointer.
func (m *TFModule) BuildDeps() error {
	for _, dep := range m.dependenciesRemoteState {
		err := project.BuildDep(m, dep)
		if err != nil {
			return err
		}
	}
	return nil
}

// PreHook return module pre hook dependency.
func (m *TFModule) PreHook() []byte {
	return m.preHook
}

// Apply module.
func (m *TFModule) Apply() error {
	rn, err := executor.NewBashRunner(m.codeDir)
	if err != nil {
		return err
	}
	rn.LogLabels = []string{
		m.InfraName(),
		m.Name(),
		"apply",
	}
	var cmd = ""
	if m.preHook != nil {
		cmd = "./pre_hook.sh && "
	}
	cmd += "terraform init && terraform apply -auto-approve"
	if m.postHook != nil {
		cmd += " && ./post_hook.sh"
	}
	_, errMsg, err := rn.Run(cmd)
	if err != nil {
		return fmt.Errorf("err: %v, error output:\n %v", err.Error(), string(errMsg))
	}
	return nil
}

// Plan module.
func (m *TFModule) Plan() error {
	rn, err := executor.NewBashRunner(m.codeDir)
	if err != nil {
		return err
	}
	rn.LogLabels = []string{
		m.InfraName(),
		m.Name(),
		"plan",
	}
	var cmd = ""
	cmd += "terraform init && terraform plan"

	_, errMsg, err := rn.Run(cmd)
	if err != nil {
		return fmt.Errorf("err: %v, error output:\n %v", err.Error(), string(errMsg))
	}
	return nil
}

// Destroy module.
func (m *TFModule) Destroy() error {
	rn, err := executor.NewBashRunner(m.codeDir)
	if err != nil {
		return err
	}
	rn.LogLabels = []string{
		m.InfraName(),
		m.Name(),
		"destroy",
	}
	var cmd = ""
	if m.preHook != nil {
		cmd = "./pre_hook.sh && "
	}
	cmd += "terraform init && terraform destroy -auto-approve"

	if m.postHook != nil {
		cmd += " && ./post_hook.sh"
	}

	_, errMsg, err := rn.Run(cmd)
	if err != nil {
		return fmt.Errorf("err: %v, error output:\n %v", err.Error(), string(errMsg))
	}
	return nil
}

// Key return uniq module index (string key for maps).
func (m *TFModule) Key() string {
	return fmt.Sprintf("%v.%v", m.InfraName(), m.name)
}
