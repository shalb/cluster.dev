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
type StateConfigSpec struct {
	CreateFiles   bool
	ApplyConf     bool
	DestroyConf   bool
	InitConf      bool
	PlanConf      bool
	Hooks         bool
	Env           bool
	WorkDir       bool
	GetOutputConf bool
}

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
	CreateFiles      FilesListT                           `yaml:"create_files,omitempty" json:"create_files,omitempty"`
	InitConf         OperationConfig                      `yaml:"init,omitempty" json:"init,omitempty"`
	ApplyConf        OperationConfig                      `yaml:"apply" json:"apply"`
	PlanConf         OperationConfig                      `yaml:"plan,omitempty" json:"plan,omitempty"`
	DestroyConf      OperationConfig                      `yaml:"destroy" json:"destroy,omitempty"`
	GetOutputsConf   OutputsConfigSpec                    `yaml:"outputs,omitempty" json:"outputs_config,omitempty"`
	OutputParsers    map[string]OutputParser              `yaml:"-" json:"-"`
	Applied          bool                                 `yaml:"-" json:"-"`
	PreHook          *HookSpec                            `yaml:"-" json:"pre_hook,omitempty"`
	PostHook         *HookSpec                            `yaml:"-" json:"post_hook,omitempty"`
	Kind             string                               `yaml:"-" json:"type"`
}

// WasApplied return true if unit's method Apply was runned.
func (m *Unit) WasApplied() bool {
	return m.Applied
}

// Markers returns list of the unit's markers.
func (m *Unit) Markers() map[string]interface{} {
	return m.UnitMarkers
}

// ReadConfig reads unit spec (unmarshaled YAML) and init the unit.
func (u *Unit) ReadConfig(spec map[string]interface{}, stack *project.Stack) error {
	if stack == nil {
		return fmt.Errorf("read shell unit: empty stack or project")
	}
	u.OutputParsers = map[string]OutputParser{
		"json":      u.JSONOutputParser,
		"regexp":    u.RegexOutputParser,
		"separator": u.SeparatorOutputParser,
	}
	// Process dependencies.
	u.UnitMarkers = make(map[string]interface{})
	u.StackPtr = stack
	u.ProjectPtr = stack.ProjectPtr
	u.SpecRaw = spec
	u.Kind = u.KindKey()
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
	_, ok = u.OutputParsers[u.GetOutputsConf.Type]
	// if !ok {
	// 	log.Debugf("Parsers: %+v", u.OutputParsers)
	// 	return fmt.Errorf("read unit: outputs config: unknown parser type '%v'", u.GetOutputsConf.Type)

	// }
	err = utils.JSONCopy(u, u.StatePtr)
	if err != nil {
		return fmt.Errorf("read unit '%v': create state: %w", u.Name(), err)
	}
	return nil
}

// ExpectedOutputs returns expected outputs of the unit.
func (m *Unit) ExpectedOutputs() map[string]*project.DependencyOutput {
	if m.Outputs == nil {
		m.Outputs = make(map[string]*project.DependencyOutput)
	}
	return m.Outputs
}

// Name return unit name.
func (m *Unit) Name() string {
	return m.MyName
}

// Stack return ptr to unit stack.
func (m *Unit) Stack() *project.Stack {
	return m.StackPtr
}

// ApplyOutput return output of unit applying.
func (m *Unit) ApplyOutput() []byte {
	return m.OutputRaw
}

// Project return ptr to unit project.
func (m *Unit) Project() *project.Project {
	return m.ProjectPtr
}

// StackName return unit stack name.
func (m *Unit) StackName() string {
	return m.StackPtr.Name
}

// Backend return unit backend.
func (m *Unit) Backend() project.Backend {
	return m.StackPtr.Backend
}

// Dependencies return slice of unit dependencies.
func (m *Unit) Dependencies() *[]*project.DependencyOutput {
	return &m.DependenciesList
}

// Init runs init procedure for unit.
func (m *Unit) Init() error {
	_, err := m.runCommands(m.InitConf, "init")
	return err
}

// Apply runs unit apply procedure.
func (m *Unit) Apply() error {

	var err error

	applyCommands := OperationConfig{}
	if m.PreHook != nil && m.PreHook.OnApply {
		applyCommands.Commands = append(applyCommands.Commands, "./pre_hook.sh")
	}
	applyCommands.Commands = append(applyCommands.Commands, m.ApplyConf.Commands...)
	if m.PostHook != nil && m.PostHook.OnApply {
		applyCommands.Commands = append(applyCommands.Commands, "./post_hook.sh")
	}
	err = project.ScanMarkers(m.Env, project.OutputsScanner, m)
	if err != nil {
		return err
	}
	err = project.ScanMarkers(m.ApplyConf.Commands, project.OutputsScanner, m)
	if err != nil {
		return err
	}
	m.OutputRaw, err = m.runCommands(applyCommands, "apply")
	if err != nil {
		return fmt.Errorf("apply unit '%v': %w", m.Key(), err)
	}
	// Get outputs.
	if m.GetOutputsConf.Command != "" {
		cmdConf := OperationConfig{
			Commands: []interface{}{
				m.GetOutputsConf.Command,
			},
		}
		m.OutputRaw, err = m.runCommands(cmdConf, "retriving outputs")
		if err != nil {
			return fmt.Errorf("retriving unit '%v' outputs: %w", m.Key(), err)
		}
	}

	if len(m.Outputs) > 0 {
		var pOutputs map[string]string
		parser, exists := m.OutputParsers[m.GetOutputsConf.Type]
		if !exists {
			return fmt.Errorf("retriving unit '%v' outputs: parser %v doesn't exists", m.Key(), m.GetOutputsConf.Type)
		}
		err = parser(string(m.OutputRaw), &pOutputs)
		if err != nil {
			return fmt.Errorf("parse outputs '%v': %w", m.GetOutputsConf.Type, err)
		}
		for _, eo := range m.Outputs {
			op, exists := pOutputs[eo.Output]
			if !exists {
				return fmt.Errorf("parse outputs: unit has no output named '%v', expected by another unit", eo.Output)
			}
			eo.OutputData = op
		}
	}

	if err == nil {
		m.Applied = true
	}
	return err
}

func (m *Unit) runCommands(commandsCnf OperationConfig, name string) ([]byte, error) {

	if len(commandsCnf.Commands) == 0 {
		log.Debugf("configuration for '%v' is empty for unit '%v'. Skip.", name, m.Key())
		return nil, nil
	}
	err := m.ReplaceMarkers()
	if err != nil {
		return nil, err
	}
	rn, err := executor.NewExecutor(m.CacheDir)
	if err != nil {
		log.Debug(err.Error())
		return nil, err
	}
	for key, val := range m.Env.(map[string]interface{}) {
		rn.Env = append(rn.Env, fmt.Sprintf("%v=%v", key, val))
	}

	rn.LogLabels = []string{
		m.StackName(),
		m.Name(),
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
func (m *Unit) Plan() error {
	planCommands := OperationConfig{}
	if m.PreHook != nil && m.PreHook.OnPlan {
		planCommands.Commands = append(planCommands.Commands, "./pre_hook.sh")
	}
	planCommands.Commands = append(planCommands.Commands, m.PlanConf.Commands...)
	if m.PostHook != nil && m.PostHook.OnPlan {
		planCommands.Commands = append(planCommands.Commands, "./post_hook.sh")
	}
	_, err := m.runCommands(planCommands, "plan")
	return err
}

// Destroy unit.
func (m *Unit) Destroy() error {
	destroyCommands := OperationConfig{}
	if m.PreHook != nil && m.PreHook.OnDestroy {
		destroyCommands.Commands = append(destroyCommands.Commands, "./pre_hook.sh")
	}
	destroyCommands.Commands = append(destroyCommands.Commands, m.DestroyConf.Commands...)
	if m.PostHook != nil && m.PostHook.OnDestroy {
		destroyCommands.Commands = append(destroyCommands.Commands, "./post_hook.sh")
	}
	_, err := m.runCommands(destroyCommands, "destroy")
	return err
}

// Key return uniq unit index (stackName.unitName).
func (m *Unit) Key() string {
	return fmt.Sprintf("%v.%v", m.StackName(), m.MyName)
}

// CodeDir return path to unit code directory.
func (m *Unit) CodeDir() string {
	return m.WorkDir
}

// UpdateProjectRuntimeData update project runtime dataset, adds unit outputs.
// TODO: get unit outputs and write to project runtime dataset. Now this function is only for printer's unit interface.
func (m *Unit) UpdateProjectRuntimeData(p *project.Project) error {
	return nil
}

// ReplaceMarkers replace all templated markers with values.
func (m *Unit) ReplaceMarkers() error {
	// log.Warnf("Replacing markers...")
	err := project.ScanMarkers(m.Env, project.OutputsScanner, m)
	if err != nil {
		return err
	}
	err = project.ScanMarkers(m.ApplyConf.Commands, project.OutputsScanner, m)
	if err != nil {
		return err
	}
	err = project.ScanMarkers(m.PlanConf.Commands, project.OutputsScanner, m)
	if err != nil {
		return err
	}
	err = project.ScanMarkers(m.DestroyConf.Commands, project.OutputsScanner, m)
	if err != nil {
		return err
	}
	return nil
}

// KindKey returns unit kind.
func (m *Unit) KindKey() string {
	return "shell"
}

// RequiredUnits list of dependencies in map representation.
func (m *Unit) RequiredUnits() map[string]project.Unit {
	res := make(map[string]project.Unit)
	for _, dep := range m.DependenciesList {
		res[dep.Unit.Key()] = m.Project().Units[dep.Unit.Key()]
	}
	return res
}

func (m *Unit) readDeps(depsData interface{}) ([]*project.DependencyOutput, error) {
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
			infNm = m.StackName()
		}
		res = append(res, &project.DependencyOutput{
			StackName: infNm,
			UnitName:  splDep[1],
		})
		log.Debugf("Dependency added: %v --> %v.%v", m.Key(), infNm, splDep[1])
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
