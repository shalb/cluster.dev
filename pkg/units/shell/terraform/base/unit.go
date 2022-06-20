package base

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/apex/log"
	"github.com/shalb/cluster.dev/pkg/config"
	"github.com/shalb/cluster.dev/pkg/executor"
	"github.com/shalb/cluster.dev/pkg/project"
	"github.com/shalb/cluster.dev/pkg/units/shell/common"
)

// const remoteStateLinkType = "RemoteStateMarkers"

// RemoteStateLinkType - name of markers category for remote states
const RemoteStateLinkType = "RemoteStateMarkers"

var terraformBin = "terraform"

type RequiredProvider struct {
	Source  string `json:"source"`
	Version string `json:"version"`
}

// Unit describe cluster.dev unit to deploy/destroy terraform modules.
type Unit struct {
	common.Unit
	// StatePtr          *Unit                       `yaml:"-" json:"-"`
	Providers         interface{}                 `yaml:"-" json:"providers,omitempty"`
	RequiredProviders map[string]RequiredProvider `yaml:"-" json:"required_providers,omitempty"`
	InitDone          bool                        `yaml:"-" json:"-"` // True if unit was initted in this session.
	StateData         interface{}                 `yaml:"-" json:"-"`
}

func (u *Unit) AddRequiredProvider(name, source, version string) {
	if u.RequiredProviders == nil {
		u.RequiredProviders = make(map[string]RequiredProvider)
	}
	u.RequiredProviders[name] = RequiredProvider{
		Version: version,
		Source:  source,
	}
}

func (u *Unit) fillShellUnit() {
	// Check if CDEV_TF_BINARY is set to change terraform binary name.
	envTfBin, exists := os.LookupEnv("CDEV_TF_BINARY")
	if exists {
		terraformBin = envTfBin
	}
	u.InitConf = &common.OperationConfig{
		Commands: []interface{}{
			fmt.Sprintf("%[1]s init", terraformBin),
		},
	}
	u.ApplyConf = &common.OperationConfig{
		Commands: []interface{}{
			fmt.Sprintf("%s apply -auto-approve", terraformBin),
		},
	}
	u.DestroyConf = &common.OperationConfig{
		Commands: []interface{}{
			fmt.Sprintf("%s destroy -auto-approve", terraformBin),
		},
	}
	u.PlanConf = &common.OperationConfig{
		Commands: []interface{}{
			fmt.Sprintf("%s plan", terraformBin),
		},
	}
	u.GetOutputsConf = &common.OutputsConfigSpec{
		Command: fmt.Sprintf("%s output -json", terraformBin),
		Type:    "terraform",
	}
	u.OutputParsers["terraform"] = TerraformJSONParser
}

func (u *Unit) ReadConfig(spec map[string]interface{}, stack *project.Stack) error {
	u.fillShellUnit()
	providers, exists := spec["providers"]
	if exists {
		u.Providers = providers
	}
	u.CacheDir = filepath.Join(u.Project().CodeCacheDir, u.Key())
	u.InitDone = false
	u.Env.(map[string]interface{})["TF_PLUGIN_CACHE_DIR"] = config.Global.PluginsCacheDir
	//err := utils.JSONCopy(m, m.StatePtr)
	return nil
}

// Init unit.
func (u *Unit) Init() error {
	u.ProjectPtr.InitLock.Lock()
	defer u.ProjectPtr.InitLock.Unlock()
	err := u.Unit.Init()
	if err != nil {
		return err
	}
	u.InitDone = true
	return nil
}

// Apply unit.
func (u *Unit) Apply() error {
	if !u.InitDone {
		if err := u.Init(); err != nil {
			return err
		}
	}
	return u.Unit.Apply()
}

// Plan unit.
func (u *Unit) Plan() error {
	if !u.InitDone {
		if err := u.Init(); err != nil {
			return err
		}
	}
	return u.Unit.Plan()
}

// Destroy unit.
func (u *Unit) Destroy() error {
	if !u.InitDone {
		if err := u.Init(); err != nil {
			return err
		}
	}
	return u.Unit.Destroy()
}

// Output unit.
func (u *Unit) Output() (string, error) {
	rn, err := executor.NewExecutor(u.CacheDir)
	if err != nil {
		log.Debug(err.Error())
		return "", err
	}
	rn.Env = append(rn.Env, fmt.Sprintf("TF_PLUGIN_CACHE_DIR=%v", config.Global.PluginsCacheDir))
	rn.LogLabels = []string{
		u.StackName(),
		u.Name(),
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
