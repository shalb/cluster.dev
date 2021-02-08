package common

import (
	"fmt"

	"github.com/apex/log"
	"github.com/shalb/cluster.dev/pkg/config"
	"github.com/shalb/cluster.dev/pkg/executor"
	"github.com/shalb/cluster.dev/pkg/project"
)

// moduleTypeKeyTf - string representation of this module type.
const moduleTypeKeyTf = "terraform"
const moduleTypeKeyKubernetes = "kubernetes"
const moduleTypeKeyHelm = "helm"
const remoteStateMarkerName = "RemoteStateMarkers"
const insertYAMLMarkerName = "insertYAMLMarkers"

// Module describe cluster.dev module to deploy/destroy terraform modules.
type Module struct {
	infraPtr        *project.Infrastructure
	projectPtr      *project.Project
	backendPtr      project.Backend
	name            string
	dependencies    []*project.Dependency
	expectedOutputs map[string]bool
	preHook         []byte
	postHook        []byte
	codeDir         string
	FilesList       map[string][]byte
	providers       interface{}
}

func (m *Module) ReadConfigCommon(spec map[string]interface{}, infra *project.Infrastructure) error {
	mName, ok := spec["name"]
	if !ok {
		return fmt.Errorf("Incorrect module name")
	}
	var modDeps []*project.Dependency
	var err error
	dependsOn, ok := spec["depends_on"]
	if ok {
		modDeps, err = readDeps(dependsOn, infra)
		if err != nil {
			log.Debug(err.Error())
			return err
		}
	}
	bPtr, exists := infra.ProjectPtr.Backends[infra.BackendName]
	if !exists {
		return fmt.Errorf("Backend '%s' not found, infra: '%s'", infra.BackendName, infra.Name)
	}

	m.infraPtr = infra
	m.projectPtr = infra.ProjectPtr
	m.name = mName.(string)
	m.dependencies = modDeps
	m.backendPtr = bPtr
	m.expectedOutputs = map[string]bool{}
	m.FilesList = map[string][]byte{}

	if err != nil {
		log.Debug(err.Error())
		return err
	}

	modPreHook, ok := spec["pre_hook"]
	if ok {
		m.preHook, err = readHook(modPreHook, "pre_hook")
		if err != nil {
			log.Debug(err.Error())
			return err
		}
	}
	modPostHook, ok := spec["post_hook"]
	if ok {
		m.postHook, err = readHook(modPostHook, "post_hook")
		if err != nil {
			log.Debug(err.Error())
			return err
		}
	}
	providers, exists := spec["providers"]
	if exists {
		m.providers = providers
	}
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
	return fmt.Errorf("internal error")
}

// Dependencies return slice of module dependencies.
func (m *Module) Dependencies() *[]*project.Dependency {
	return &m.dependencies
}

// Apply module.
func (m *Module) Apply() error {
	rn, err := executor.NewBashRunner(m.codeDir)
	if err != nil {
		log.Debug(err.Error())
		return err
	}
	rn.Env = append(rn.Env, fmt.Sprintf("TF_PLUGIN_CACHE_DIR=%v", config.Global.PluginsCacheDir))
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
func (m *Module) Plan() error {
	rn, err := executor.NewBashRunner(m.codeDir)
	if err != nil {
		log.Debug(err.Error())
		return err
	}
	rn.Env = append(rn.Env, fmt.Sprintf("TF_PLUGIN_CACHE_DIR=%v", config.Global.PluginsCacheDir))
	rn.LogLabels = []string{
		m.InfraName(),
		m.Name(),
		"plan",
	}
	var cmd = ""

	cmd += "terraform init && terraform plan"

	planOutput, errMsg, err := rn.Run(cmd)
	if err != nil {
		log.Debug(err.Error())
		return fmt.Errorf("err: %v, error output:\n %v", err.Error(), string(errMsg))
	}
	log.Infof("Module '%v', plan output:\v%v", m.Key(), string(planOutput))
	return nil
}

// Destroy module.
func (m *Module) Destroy() error {
	rn, err := executor.NewBashRunner(m.codeDir)
	rn.Env = append(rn.Env, fmt.Sprintf("TF_PLUGIN_CACHE_DIR=%v", config.Global.PluginsCacheDir))
	if err != nil {
		log.Debug(err.Error())
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
func (m *Module) Key() string {
	return fmt.Sprintf("%v.%v", m.InfraName(), m.name)
}
