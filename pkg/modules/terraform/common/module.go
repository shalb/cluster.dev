package common

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/apex/log"
	"github.com/shalb/cluster.dev/pkg/config"
	"github.com/shalb/cluster.dev/pkg/executor"
	"github.com/shalb/cluster.dev/pkg/project"
)

const remoteStateMarkerName = "RemoteStateMarkers"

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

// Unit describe cluster.dev unit to deploy/destroy terraform units.
type Unit struct {
	stackPtr          *project.Stack
	projectPtr        *project.Project
	backendPtr        project.Backend
	name              string
	dependencies      []*project.DependencyOutput
	expectedOutputs   map[string]*project.DependencyOutput
	preHook           *hookSpec
	postHook          *hookSpec
	codeDir           string
	filesList         map[string][]byte
	providers         interface{}
	specRaw           map[string]interface{}
	markers           map[string]interface{}
	applyOutput       []byte
	requiredProviders map[string]RequiredProvider
	applied           bool
}

// WasApplied return true if unit's method Apply was runned.
func (m *Unit) WasApplied() bool {
	return m.applied
}

func (m *Unit) AddRequiredProvider(name, source, version string) {
	if m.requiredProviders == nil {
		m.requiredProviders = make(map[string]RequiredProvider)
	}
	m.requiredProviders[name] = RequiredProvider{
		Version: version,
		Source:  source,
	}
}

func (m *Unit) Markers() map[string]interface{} {
	return m.markers
}

func (m *Unit) FilesList() map[string][]byte {
	return m.filesList
}

func (m *Unit) ReadConfig(spec map[string]interface{}, stack *project.Stack) error {
	// Check if CDEV_TF_BINARY is set to change terraform binary name.
	envTfBin, exists := os.LookupEnv("CDEV_TF_BINARY")
	if exists {
		terraformBin = envTfBin
	}
	mName, ok := spec["name"]
	if !ok {
		return fmt.Errorf("Incorrect unit name")
	}

	m.stackPtr = stack
	m.projectPtr = stack.ProjectPtr
	m.name = mName.(string)
	m.expectedOutputs = make(map[string]*project.DependencyOutput)
	m.filesList = make(map[string][]byte)
	m.specRaw = spec
	m.applied = false
	m.markers = make(map[string]interface{})

	// Process dependencies.
	var modDeps []*project.DependencyOutput
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
	bPtr, exists := stack.ProjectPtr.Backends[stack.BackendName]
	if !exists {
		return fmt.Errorf("Backend '%s' not found, stack: '%s'", stack.BackendName, stack.Name)
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
	m.codeDir = filepath.Join(m.ProjectPtr().CodeCacheDir, m.Key())
	return nil
}

func (m *Unit) ExpectedOutputs() map[string]*project.DependencyOutput {
	return m.expectedOutputs
}

// Name return unit name.
func (m *Unit) Name() string {
	return m.name
}

// StackPtr return ptr to unit stack.
func (m *Unit) StackPtr() *project.Stack {
	return m.stackPtr
}

// ApplyOutput return output of last unit applying.
func (m *Unit) ApplyOutput() []byte {
	return m.applyOutput
}

// ProjectPtr return ptr to unit project.
func (m *Unit) ProjectPtr() *project.Project {
	return m.projectPtr
}

// StackName return unit stack name.
func (m *Unit) StackName() string {
	return m.stackPtr.Name
}

// Backend return unit backend.
func (m *Unit) Backend() project.Backend {
	return m.stackPtr.Backend
}

// Dependencies return slice of unit dependencies.
func (m *Unit) Dependencies() *[]*project.DependencyOutput {
	return &m.dependencies
}

func (m *Unit) Init() error {
	rn, err := executor.NewExecutor(m.codeDir)
	if err != nil {
		log.Debug(err.Error())
		return err
	}
	rn.Env = append(rn.Env, fmt.Sprintf("TF_PLUGIN_CACHE_DIR=%v", config.Global.PluginsCacheDir))
	rn.LogLabels = []string{
		m.StackName(),
		m.Name(),
		"init",
	}

	var cmd = ""
	cmd += fmt.Sprintf("%[1]s init", terraformBin)
	var errMsg []byte
	m.projectPtr.InitLock.Lock()
	defer m.projectPtr.InitLock.Unlock()
	m.applyOutput, errMsg, err = rn.Run(cmd)
	if err != nil {
		if len(errMsg) > 1 {
			return fmt.Errorf("%v, error output:\n %v", err.Error(), string(errMsg))
		}
	}
	return err
}

func (m *Unit) Apply() error {
	rn, err := executor.NewExecutor(m.codeDir)
	if err != nil {
		log.Debug(err.Error())
		return err
	}
	rn.Env = append(rn.Env, fmt.Sprintf("TF_PLUGIN_CACHE_DIR=%v", config.Global.PluginsCacheDir))
	rn.LogLabels = []string{
		m.StackName(),
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
		if len(errMsg) > 1 {
			return fmt.Errorf("%v, error output:\n %v", err.Error(), string(errMsg))
		}
	}
	if err == nil {
		m.applied = true
	}
	return err
}

// Output unit.
func (m *Unit) Output() (string, error) {
	rn, err := executor.NewExecutor(m.codeDir)
	if err != nil {
		log.Debug(err.Error())
		return "", err
	}
	rn.Env = append(rn.Env, fmt.Sprintf("TF_PLUGIN_CACHE_DIR=%v", config.Global.PluginsCacheDir))
	rn.LogLabels = []string{
		m.StackName(),
		m.Name(),
		"plan",
	}
	var cmd = ""
	cmd += fmt.Sprintf("%s output", terraformBin)

	var errMsg []byte
	res, errMsg, err := rn.Run(cmd)

	if err != nil {
		if len(errMsg) > 1 {
			return "", fmt.Errorf("%v, error output:\n %v", err.Error(), string(errMsg))
		}
	}
	return string(res), err
}

// Plan unit.
func (m *Unit) Plan() error {
	rn, err := executor.NewExecutor(m.codeDir)
	if err != nil {
		log.Debug(err.Error())
		return err
	}
	rn.Env = append(rn.Env, fmt.Sprintf("TF_PLUGIN_CACHE_DIR=%v", config.Global.PluginsCacheDir))
	rn.LogLabels = []string{
		m.StackName(),
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
		if len(errMsg) > 1 {
			return fmt.Errorf("%v, error output:\n %v", err.Error(), string(errMsg))
		}
		return err
	}
	fmt.Printf("%v\n", string(planOutput))
	return nil
}

// Destroy unit.
func (m *Unit) Destroy() error {
	rn, err := executor.NewExecutor(m.codeDir)
	rn.Env = append(rn.Env, fmt.Sprintf("TF_PLUGIN_CACHE_DIR=%v", config.Global.PluginsCacheDir))
	if err != nil {
		log.Debug(err.Error())
		return err
	}
	rn.LogLabels = []string{
		m.StackName(),
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
		if len(errMsg) > 1 {
			return fmt.Errorf("%v, error output:\n %v", err.Error(), string(errMsg))
		}
	}
	return err
}

// Key return uniq unit index (string key for maps).
func (m *Unit) Key() string {
	return fmt.Sprintf("%v.%v", m.StackName(), m.name)
}

// CodeDir return path to unit code directory.
func (m *Unit) CodeDir() string {
	return m.codeDir
}

// UpdateProjectRuntimeData update project runtime dataset, adds unit outputs.
func (m *Unit) UpdateProjectRuntimeData(p *project.Project) error {
	return nil
}

// ReplaceMarkers replace all templated markers with values.
func (m *Unit) ReplaceMarkers(inheritedUnit project.Unit) error {
	if m.preHook != nil {
		err := project.ScanMarkers(&m.preHook.Command, m.RemoteStatesScanner, inheritedUnit)
		if err != nil {
			return err
		}
	}
	if m.postHook != nil {
		err := project.ScanMarkers(&m.postHook.Command, m.RemoteStatesScanner, inheritedUnit)
		if err != nil {
			return err
		}
	}
	return nil
}
