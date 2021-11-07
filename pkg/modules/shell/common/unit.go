package common

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/apex/log"
	"github.com/shalb/cluster.dev/pkg/config"
	"github.com/shalb/cluster.dev/pkg/executor"
	"github.com/shalb/cluster.dev/pkg/project"
	"github.com/shalb/cluster.dev/pkg/utils"
	"gopkg.in/yaml.v3"
)

// HookSpec describes pre/post hooks configuration in unit YAML.
type HookSpec struct {
	Command   string `json:"command"`
	OnDestroy bool   `yaml:"on_destroy,omitempty" json:"on_destroy,omitempty"`
	OnApply   bool   `yaml:"on_apply,omitempty" json:"on_apply,omitempty"`
	OnPlan    bool   `yaml:"on_plan,omitempty" json:"on_plan,omitempty"`
}

// OperationConfig type that describe apply, plan and destroy operations.
type OperationConfig struct {
	Commands []interface{} `yaml:"commands" json:"commands"`
}

// OutputsConfigSpec describe how to retrive parse unit outputs.
type OutputsConfigSpec struct {
	Command   string `yaml:"command,omitempty" json:"command,omitempty"`
	Type      string `yaml:"type" json:"type"`
	Regexp    string `yaml:"regexp,omitempty" json:"regexp,omitempty"`
	Separator string `yaml:"separator,omitempty" json:"separator,omitempty"`
}

// type StateConfigFileSpec struct {
// 	Mask      string `yaml:"mask"`
// 	Dir       string `yaml:"dir"`
// 	Recursive bool   `yaml:"recursive"`
// }

// StateConfigSpec describes what data to save to the state.
// type StateConfigSpec struct {
// 	CreateFiles   bool
// 	ApplyConf     bool
// 	DestroyConf   bool
// 	InitConf      bool
// 	PlanConf      bool
// 	Hooks         bool
// 	Env           bool
// 	WorkDir       bool
// 	GetOutputConf bool
// }

// OutputParser represents function for parsing unit output.
type OutputParser func(string, interface{}) error

// Unit describe cluster.dev shell unit.
type Unit struct {
	StatePtr         *Unit                                `yaml:"-" json:"-"`
	StackPtr         *project.Stack                       `yaml:"-" json:"-"`
	ProjectPtr       *project.Project                     `yaml:"-" json:"-"`
	DependenciesList []*project.DependencyOutput          `yaml:"-" json:"dependencies,omitempty"`
	Outputs          map[string]*project.DependencyOutput `yaml:"-" json:"outputs,omitempty"`
	SpecRaw          map[string]interface{}               `yaml:"-" json:"-"`
	UnitMarkers      map[string]interface{}               `yaml:"-" json:"-" json:"markers,omitempty"`
	OutputRaw        []byte                               `yaml:"-" json:"-"`
	CacheDir         string                               `yaml:"-" json:"-"`
	MyName           string                               `yaml:"name" json:"name"`
	WorkDir          string                               `yaml:"work_dir,omitempty" json:"work_dir,omitempty"`
	Env              interface{}                          `yaml:"env,omitempty" json:"env,omitempty"`
	CreateFiles      *FilesListT                          `yaml:"create_files,omitempty" json:"create_files,omitempty"`
	InitConf         *OperationConfig                     `yaml:"init,omitempty" json:"init,omitempty"`
	ApplyConf        *OperationConfig                     `yaml:"apply" json:"apply"`
	PlanConf         *OperationConfig                     `yaml:"plan,omitempty" json:"plan,omitempty"`
	DestroyConf      *OperationConfig                     `yaml:"destroy" json:"destroy,omitempty"`
	GetOutputsConf   *OutputsConfigSpec                   `yaml:"outputs,omitempty" json:"outputs_config,omitempty"`
	OutputParsers    map[string]OutputParser              `yaml:"-" json:"-"`
	Applied          bool                                 `yaml:"-" json:"-"`
	PreHook          *HookSpec                            `yaml:"-" json:"pre_hook,omitempty"`
	PostHook         *HookSpec                            `yaml:"-" json:"post_hook,omitempty"`
	UnitKind         string                               `yaml:"-" json:"type"`
	BackendPtr       *project.Backend                     `yaml:"-" json:"-"`
	BackendName      string                               `yaml:"-" json:"backend_name"`
}

// WasApplied return true if unit's method Apply was runned.
func (u *Unit) WasApplied() bool {
	return u.Applied
}

// Markers returns list of the unit's markers.
func (u *Unit) Markers() map[string]interface{} {
	return u.UnitMarkers
}

// ReadConfig reads unit spec (unmarshaled YAML) and init the unit.
func (u *Unit) ReadConfig(spec map[string]interface{}, stack *project.Stack) error {
	if stack == nil {
		return fmt.Errorf("read shell unit: empty stack or project")
	}
	u.StackPtr = stack
	u.ProjectPtr = stack.ProjectPtr
	u.SpecRaw = spec
	var modDeps []*project.DependencyOutput
	var err error
	dependsOn, ok := spec["depends_on"]
	if ok {
		modDeps, err = u.readDeps(dependsOn)
		if err != nil {
			log.Debug(err.Error())
			return err
		}
	}
	u.DependenciesList = modDeps
	err = utils.YAMLInterfaceToType(spec, u)
	if err != nil {
		return err
	}
	if u.WorkDir != "" {
		u.WorkDir = filepath.Join(config.Global.WorkingDir, u.StackPtr.TemplateDir, u.WorkDir)
		isDir, err := utils.CheckDir(u.WorkDir)
		if err != nil {
			return fmt.Errorf("read unit '%v': check working dir: %v", u.Name(), err.Error())
		}
		if !isDir {
			return fmt.Errorf("read unit: check working dir: '%v' is not a directory", u.WorkDir)
		}
	}
	// Check and set backend.
	bk, exists := stack.ProjectPtr.Backends[stack.BackendName]
	if !exists {
		return fmt.Errorf("Backend '%s' not found, stack: '%s'", stack.BackendName, stack.Name)
	}
	u.BackendPtr = &bk
	// Process hooks.
	modPreHook, ok := spec["pre_hook"]
	if ok {
		u.PreHook, err = readHook(modPreHook, "pre_hook")
		if err != nil {
			return fmt.Errorf("read unit: pre_hook: %w", err)
		}

	}
	modPostHook, ok := spec["post_hook"]
	if ok {
		u.PostHook, err = readHook(modPostHook, "post_hook")
		if err != nil {
			return fmt.Errorf("read unit: post_hook: %w", err)
		}
	}
	u.CacheDir = filepath.Join(u.Project().CodeCacheDir, u.Key())
	err = utils.JSONCopy(u, u.StatePtr)
	if err != nil {
		return fmt.Errorf("read unit '%v': create state: %w", u.Name(), err)
	}
	return nil
}

// ExpectedOutputs returns expected outputs of the unit.
func (u *Unit) ExpectedOutputs() map[string]*project.DependencyOutput {
	if u.Outputs == nil {
		u.Outputs = make(map[string]*project.DependencyOutput)
	}
	return u.Outputs
}

// Name return unit name.
func (u *Unit) Name() string {
	return u.MyName
}

// Stack return ptr to unit stack.
func (u *Unit) Stack() *project.Stack {
	return u.StackPtr
}

// ApplyOutput return output of unit applying.
func (u *Unit) ApplyOutput() []byte {
	return u.OutputRaw
}

// Project return ptr to unit project.
func (u *Unit) Project() *project.Project {
	return u.ProjectPtr
}

// StackName return unit stack name.
func (u *Unit) StackName() string {
	return u.StackPtr.Name
}

// Backend return unit backend.
func (u *Unit) Backend() project.Backend {
	return u.StackPtr.Backend
}

// Dependencies return slice of unit dependencies.
func (u *Unit) Dependencies() *[]*project.DependencyOutput {
	return &u.DependenciesList
}

// Init runs init procedure for unit.
func (u *Unit) Init() error {
	_, err := u.runCommands(*u.InitConf, "init")
	return err
}

// Apply runs unit apply procedure.
func (u *Unit) Apply() error {

	var err error

	applyCommands := OperationConfig{}
	if u.PreHook != nil && u.PreHook.OnApply {
		applyCommands.Commands = append(applyCommands.Commands, "./pre_hook.sh")
	}
	applyCommands.Commands = append(applyCommands.Commands, u.ApplyConf.Commands...)
	if u.PostHook != nil && u.PostHook.OnApply {
		applyCommands.Commands = append(applyCommands.Commands, "./post_hook.sh")
	}
	err = project.ScanMarkers(u.Env, project.OutputsScanner, u)
	if err != nil {
		return err
	}
	err = project.ScanMarkers(u.ApplyConf.Commands, project.OutputsScanner, u)
	if err != nil {
		return err
	}
	u.OutputRaw, err = u.runCommands(applyCommands, "apply")
	if err != nil {
		return fmt.Errorf("apply unit '%v': %w", u.Key(), err)
	}
	// Get outputs.
	if u.GetOutputsConf.Command != "" {
		cmdConf := OperationConfig{
			Commands: []interface{}{
				u.GetOutputsConf.Command,
			},
		}
		u.OutputRaw, err = u.runCommands(cmdConf, "retriving outputs")
		if err != nil {
			return fmt.Errorf("retriving unit '%v' outputs: %w", u.Key(), err)
		}
	}

	if len(u.Outputs) > 0 {
		var pOutputs map[string]string
		parser, exists := u.OutputParsers[u.GetOutputsConf.Type]
		if !exists {
			return fmt.Errorf("retriving unit '%v' outputs: parser %v doesn't exists", u.Key(), u.GetOutputsConf.Type)
		}
		err = parser(string(u.OutputRaw), &pOutputs)
		if err != nil {
			return fmt.Errorf("parse outputs '%v': %w", u.GetOutputsConf.Type, err)
		}
		for _, eo := range u.Outputs {
			op, exists := pOutputs[eo.Output]
			if !exists {
				return fmt.Errorf("parse outputs: unit has no output named '%v', expected by another unit", eo.Output)
			}
			eo.OutputData = op
		}
	}

	if err == nil {
		u.Applied = true
	}
	return err
}

func (u *Unit) runCommands(commandsCnf OperationConfig, name string) ([]byte, error) {

	if len(commandsCnf.Commands) == 0 {
		log.Debugf("configuration for '%v' is empty for unit '%v'. Skip.", name, u.Key())
		return nil, nil
	}
	err := u.ReplaceMarkers()
	if err != nil {
		return nil, err
	}
	rn, err := executor.NewExecutor(u.CacheDir)
	if err != nil {
		log.Debug(err.Error())
		return nil, err
	}

	for key, val := range u.Env.(map[string]interface{}) {
		rn.Env = append(rn.Env, fmt.Sprintf("%v=%v", key, val))
	}

	rn.LogLabels = []string{
		u.StackName(),
		u.Name(),
		name,
	}
	var errMsg []byte

	var cmd string

	for i, c := range commandsCnf.Commands {
		cmd += fmt.Sprintf("%v", c)
		if i < len(commandsCnf.Commands)-1 {
			cmd += "\n"
		}
	}
	otp, errMsg, err := rn.Run(cmd)
	if len(errMsg) > 1 {
		log.Errorf("%v", string(errMsg))
		return otp, fmt.Errorf("%v, error output:\n %v", err.Error(), string(errMsg))
	}
	return otp, err
}

// Plan unit.
func (u *Unit) Plan() error {
	planCommands := OperationConfig{}
	if u.PreHook != nil && u.PreHook.OnPlan {
		planCommands.Commands = append(planCommands.Commands, "./pre_hook.sh")
	}
	planCommands.Commands = append(planCommands.Commands, u.PlanConf.Commands...)
	if u.PostHook != nil && u.PostHook.OnPlan {
		planCommands.Commands = append(planCommands.Commands, "./post_hook.sh")
	}
	_, err := u.runCommands(planCommands, "plan")
	return err
}

// Destroy unit.
func (u *Unit) Destroy() error {
	destroyCommands := OperationConfig{}
	if u.PreHook != nil && u.PreHook.OnDestroy {
		destroyCommands.Commands = append(destroyCommands.Commands, "./pre_hook.sh")
	}
	destroyCommands.Commands = append(destroyCommands.Commands, u.DestroyConf.Commands...)
	if u.PostHook != nil && u.PostHook.OnDestroy {
		destroyCommands.Commands = append(destroyCommands.Commands, "./post_hook.sh")
	}
	_, err := u.runCommands(destroyCommands, "destroy")
	return err
}

// Key return uniq unit index (stackName.unitName).
func (u *Unit) Key() string {
	return fmt.Sprintf("%v.%v", u.StackName(), u.MyName)
}

// CodeDir return path to unit code directory.
func (u *Unit) CodeDir() string {
	return u.WorkDir
}

// UpdateProjectRuntimeData update project runtime dataset, adds unit outputs.
// TODO: get unit outputs and write to project runtime dataset. Now this function is only for printer's unit interface.
func (u *Unit) UpdateProjectRuntimeData(p *project.Project) error {
	return nil
}

// ReplaceMarkers replace all templated markers with values.
func (u *Unit) ReplaceMarkers() error {
	// log.Warnf("Replacing markers...")
	if u.Env != nil {
		err := project.ScanMarkers(u.Env, project.OutputsScanner, u)
		if err != nil {
			return err
		}
	}

	if u.ApplyConf != nil {
		err := project.ScanMarkers(u.ApplyConf.Commands, project.OutputsScanner, u)
		if err != nil {
			return err
		}
	}

	if u.PlanConf != nil {
		err := project.ScanMarkers(u.PlanConf.Commands, project.OutputsScanner, u)
		if err != nil {
			return err
		}
	}

	if u.DestroyConf != nil {
		err := project.ScanMarkers(u.DestroyConf.Commands, project.OutputsScanner, u)
		if err != nil {
			return err
		}
	}

	if u.InitConf != nil {
		err := project.ScanMarkers(u.InitConf.Commands, project.OutputsScanner, u)
		if err != nil {
			return err
		}
	}
	if u.PreHook != nil {
		err := project.ScanMarkers(&u.PreHook.Command, project.OutputsScanner, u)
		if err != nil {
			return err
		}
	}
	if u.PostHook != nil {
		err := project.ScanMarkers(&u.PostHook.Command, project.OutputsScanner, u)
		if err != nil {
			return err
		}
	}
	return nil
}

// KindKey returns unit kind.
func (u *Unit) KindKey() string {
	return "shell"
}

// RequiredUnits list of dependencies in map representation.
func (u *Unit) RequiredUnits() map[string]project.Unit {
	res := make(map[string]project.Unit)
	for _, dep := range u.DependenciesList {
		res[dep.Unit.Key()] = u.Project().Units[dep.Unit.Key()]
	}
	return res
}

func (u *Unit) readDeps(depsData interface{}) ([]*project.DependencyOutput, error) {
	rawDepsList := []string{}
	switch depsData.(type) {
	case string:
		rawDepsList = append(rawDepsList, depsData.(string))
	case []string:
		rawDepsList = append(rawDepsList, depsData.([]string)...)
	}
	var res []*project.DependencyOutput
	for _, dep := range rawDepsList {
		splDep := strings.Split(dep, ".")
		if len(splDep) != 2 {
			return nil, fmt.Errorf("Incorrect unit dependency '%v'", dep)
		}
		infNm := splDep[0]
		if infNm == "this" {
			infNm = u.StackName()
		}
		res = append(res, &project.DependencyOutput{
			StackName: infNm,
			UnitName:  splDep[1],
		})
		log.Debugf("Dependency added: %v --> %v.%v", u.Key(), infNm, splDep[1])
	}
	return res, nil
}

func readHook(hookData interface{}, hookType string) (*HookSpec, error) {
	hook, ok := hookData.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("%s configuration error", hookType)
	}
	cmd, cmdExists := hook["command"].(string)

	if !cmdExists {
		return nil, fmt.Errorf("Error in %s config, use 'script' option", hookType)
	}

	ScriptData := HookSpec{
		Command:   "",
		OnDestroy: false,
		OnApply:   true,
		OnPlan:    false,
	}
	ymlTmp, err := yaml.Marshal(hookData)
	if err != nil {
		log.Debug(err.Error())
		return nil, err
	}
	err = yaml.Unmarshal(ymlTmp, &ScriptData)
	if err != nil {
		log.Debug(err.Error())
		return nil, err
	}
	ScriptData.Command = fmt.Sprintf("#!/usr/bin/env sh\nset -e\n\n%s", cmd)
	return &ScriptData, nil

}
