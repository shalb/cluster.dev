package common

import (
	"fmt"
	"os"

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

var terraformBin = "terraform"

type hookSpec struct {
	Command   string `json:"command"`
	OnDestroy bool   `yaml:"on_destroy,omitempty" json:"on_destroy,omitempty"`
	OnApply   bool   `yaml:"on_apply,omitempty" json:"on_apply,omitempty"`
	OnPlan    bool   `yaml:"on_plan,omitempty" json:"on_plan,omitempty"`
}

type RequiredProvider struct {
	Source  string `json:"source"`
	Version string `json:"version"`
}

// Module describe cluster.dev module to deploy/destroy terraform modules.
type Module struct {
	infraPtr          *project.Infrastructure
	projectPtr        *project.Project
	backendPtr        project.Backend
	name              string
	dependencies      []*project.Dependency
	expectedOutputs   map[string]bool
	preHook           *hookSpec
	postHook          *hookSpec
	codeDir           string
	filesList         map[string][]byte
	providers         interface{}
	specRaw           map[string]interface{}
	markers           map[string]string
	applyOutput       []byte
	requiredProviders map[string]RequiredProvider
}

func (m *Module) AddRequiredProvider(name, source, version string) {
	if m.requiredProviders == nil {
		m.requiredProviders = make(map[string]RequiredProvider)
	}
	m.requiredProviders[name] = RequiredProvider{
		Version: version,
		Source:  source,
	}
}

func (m *Module) Markers() map[string]string {
	return m.markers
}

func (m *Module) FilesList() map[string][]byte {
	return m.filesList
}

func (m *Module) ReadConfigCommon(spec map[string]interface{}, infra *project.Infrastructure) error {
	// Check if CDEV_TF_BINARY is set to change terraform binary name.
	envTfBin, exists := os.LookupEnv("CDEV_TF_BINARY")
	if exists {
		terraformBin = envTfBin
	}
	mName, ok := spec["name"]
	if !ok {
		return fmt.Errorf("Incorrect module name")
	}

	m.infraPtr = infra
	m.projectPtr = infra.ProjectPtr
	m.name = mName.(string)
	m.expectedOutputs = map[string]bool{}
	m.filesList = map[string][]byte{}
	m.specRaw = spec
	m.markers = map[string]string{}

	// Process dependencies.
	var modDeps []*project.Dependency
	var err error
	dependsOn, ok := spec["depends_on"]
	if ok {
		modDeps, err = m.readDeps(dependsOn)
		if err != nil {
			log.Debug(err.Error())
			return err
		}
	}
	m.dependencies = modDeps

	// Check and set backend.
	bPtr, exists := infra.ProjectPtr.Backends[infra.BackendName]
	if !exists {
		return fmt.Errorf("Backend '%s' not found, infra: '%s'", infra.BackendName, infra.Name)
	}
	m.backendPtr = bPtr

	// Process hooks.
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
	// Set providers.
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

// ApplyOutput return output of last module applying.
func (m *Module) ApplyOutput() []byte {
	return m.applyOutput
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

// // ReplaceMarkers replace all templated markers with values.
// func (m *Module) ReplaceMarkers() error {
// 	return fmt.Errorf("internal error")
// }

// Dependencies return slice of module dependencies.
func (m *Module) Dependencies() *[]*project.Dependency {
	return &m.dependencies
}

func (m *Module) InitDefault() error {
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
	cmd += fmt.Sprintf("%[1]s init && %[1]s apply -auto-approve", terraformBin)
	if m.postHook != nil && m.postHook.OnApply {
		cmd += " && ./post_hook.sh"
	}
	var errMsg []byte
	m.applyOutput, errMsg, err = rn.Run(cmd)
	if err != nil {
		return fmt.Errorf("err: %v, error output:\n %v", err.Error(), string(errMsg))
	}
	// log.Info(colors.LightWhiteBold.Sprint("successfully applied"))
	return nil
}

func (m *Module) ApplyDefault() error {
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
	if m.preHook != nil && m.preHook.OnApply {
		cmd = "./pre_hook.sh && "
	}
	cmd += fmt.Sprintf("%[1]s init && %[1]s apply -auto-approve", terraformBin)
	if m.postHook != nil && m.postHook.OnApply {
		cmd += " && ./post_hook.sh"
	}
	var errMsg []byte
	m.applyOutput, errMsg, err = rn.Run(cmd)
	if err != nil {
		return fmt.Errorf("err: %v, error output:\n %v", err.Error(), string(errMsg))
	}
	// log.Info(colors.LightWhiteBold.Sprint("successfully applied"))
	return nil
}

// Apply module.
func (m *Module) Apply() error {
	return m.ApplyDefault()
}

// Outputs module.
func (m *Module) Outputs() (string, error) {
	rn, err := executor.NewBashRunner(m.codeDir)
	if err != nil {
		log.Debug(err.Error())
		return "", err
	}
	rn.Env = append(rn.Env, fmt.Sprintf("TF_PLUGIN_CACHE_DIR=%v", config.Global.PluginsCacheDir))
	rn.LogLabels = []string{
		m.InfraName(),
		m.Name(),
		"plan",
	}
	var cmd = ""
	cmd += fmt.Sprintf("%s output", terraformBin)

	var errMsg []byte
	res, errMsg, err := rn.Run(cmd)
	if err != nil {
		return "", fmt.Errorf("err: %v, error output:\n %v", err.Error(), string(errMsg))
	}
	return string(res), nil
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
	if m.preHook != nil && m.preHook.OnPlan {
		cmd = "./pre_hook.sh && "
	}
	cmd += fmt.Sprintf("%[1]s init && %[1]s plan", terraformBin)

	if m.postHook != nil && m.postHook.OnPlan {
		cmd += " && ./post_hook.sh"
	}
	planOutput, errMsg, err := rn.Run(cmd)
	if err != nil {
		log.Debug(err.Error())
		return fmt.Errorf("err: %v, error output:\n %v", err.Error(), string(errMsg))
	}
	fmt.Printf("%v\n", string(planOutput))
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
	if m.preHook != nil && m.preHook.OnDestroy {
		cmd = "./pre_hook.sh && "
	}
	cmd += fmt.Sprintf("%[1]s init && %[1]s destroy -auto-approve", terraformBin)

	if m.postHook != nil && m.postHook.OnDestroy {
		cmd += " && ./post_hook.sh"
	}

	_, errMsg, err := rn.Run(cmd)
	if err != nil {
		return fmt.Errorf("err: %v, error output:\n %v", err.Error(), string(errMsg))
	}
	// log.Info(colors.LightWhiteBold.Sprint("successfully destroyed"))
	return nil
}

// Key return uniq module index (string key for maps).
func (m *Module) Key() string {
	return fmt.Sprintf("%v.%v", m.InfraName(), m.name)
}
