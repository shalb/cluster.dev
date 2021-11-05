package base

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/apex/log"
	"github.com/shalb/cluster.dev/pkg/config"
	"github.com/shalb/cluster.dev/pkg/executor"
	"github.com/shalb/cluster.dev/pkg/modules/shell/common"
	"github.com/shalb/cluster.dev/pkg/project"
)

const remoteStateMarkerName = "RemoteStateMarkers"

var terraformBin = "terraform"

type RequiredProvider struct {
	Source  string `json:"source"`
	Version string `json:"version"`
}

// Unit describe cluster.dev unit to deploy/destroy terraform modules.
type Unit struct {
	common.Unit
	BackendPtr        project.Backend             `yaml:"-" json:"-"`
	BackendName       string                      `yaml:"-" json:"backend_name"`
	Providers         interface{}                 `yaml:"providers,omitempty" json:"providers,omitempty"`
	RequiredProviders map[string]RequiredProvider `yaml:"required_providers,omitempty" json:"required_providers,omitempty"`
	initted           bool                        // True if unit was initted in this session.
}

func (m *Unit) AddRequiredProvider(name, source, version string) {
	if m.RequiredProviders == nil {
		m.RequiredProviders = make(map[string]RequiredProvider)
	}
	m.RequiredProviders[name] = RequiredProvider{
		Version: version,
		Source:  source,
	}
}

func (m *Unit) ReadConfig(spec map[string]interface{}, stack *project.Stack) error {
	// Check if CDEV_TF_BINARY is set to change terraform binary name.
	envTfBin, exists := os.LookupEnv("CDEV_TF_BINARY")
	if exists {
		terraformBin = envTfBin
	}
	m.InitConf.Commands = append(m.InitConf.Commands, fmt.Sprintf("%[1]s init", terraformBin))
	m.ApplyConf.Commands = append(m.ApplyConf.Commands, fmt.Sprintf("%s apply -auto-approve", terraformBin))
	m.PlanConf.Commands = append(m.PlanConf.Commands, fmt.Sprintf("%s plan", terraformBin))
	m.DestroyConf.Commands = append(m.DestroyConf.Commands, fmt.Sprintf("%s destroy -auto-approve", terraformBin))
	m.GetOutputsConf.Command = fmt.Sprintf("%s output -json", terraformBin)
	m.OutputParsers = make(map[string]common.OutputParser)
	m.GetOutputsConf.Type = "terraform"
	err := m.Unit.ReadConfig(spec, stack)
	if err != nil {
		return err
	}
	// Check and set backend.
	bPtr, exists := stack.ProjectPtr.Backends[stack.BackendName]
	if !exists {
		return fmt.Errorf("Backend '%s' not found, stack: '%s'", stack.BackendName, stack.Name)
	}
	m.BackendPtr = bPtr

	// Set providers.
	providers, exists := spec["providers"]
	if exists {
		m.Providers = providers
	}
	m.CacheDir = filepath.Join(m.Project().CodeCacheDir, m.Key())
	if m.Env == nil {
		m.Env = make(map[string]interface{})
	}
	m.Env.(map[string]interface{})["TF_PLUGIN_CACHE_DIR"] = config.Global.PluginsCacheDir
	m.initted = false
	return nil
}

// Init unit.
func (m *Unit) Init() error {
	m.ProjectPtr.InitLock.Lock()
	defer m.ProjectPtr.InitLock.Unlock()
	err := m.Unit.Init()
	if err != nil {
		return err
	}
	m.initted = true
	return nil
}

// Apply unit.
func (m *Unit) Apply() error {
	if !m.initted {
		if err := m.Init(); err != nil {
			return err
		}
	}
	return m.Unit.Apply()
}

// Plan unit.
func (m *Unit) Plan() error {
	if !m.initted {
		if err := m.Init(); err != nil {
			return err
		}
	}
	return m.Unit.Plan()
}

// Destroy unit.
func (m *Unit) Destroy() error {
	if !m.initted {
		if err := m.Init(); err != nil {
			return err
		}
	}
	return m.Unit.Destroy()
}

// Output unit.
func (m *Unit) Output() (string, error) {
	rn, err := executor.NewExecutor(m.CacheDir)
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

// ReplaceMarkers replace all templated markers with values.
func (m *Unit) ReplaceMarkers(inheritedUnit project.Unit) error {
	if err := m.Unit.ReplaceMarkers(); err != nil {
		return fmt.Errorf("prepare terraform unit data: %w", err)
	}
	if m.PreHook != nil {
		err := project.ScanMarkers(&m.PreHook.Command, m.RemoteStatesScanner, inheritedUnit)
		if err != nil {
			return err
		}
	}
	if m.PostHook != nil {
		err := project.ScanMarkers(&m.PostHook.Command, m.RemoteStatesScanner, inheritedUnit)
		if err != nil {
			return err
		}
	}
	return nil
}
