package common

import (
	"fmt"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"

	"github.com/apex/log"
	"github.com/shalb/cluster.dev/pkg/config"
	"github.com/shalb/cluster.dev/pkg/executor"
	"github.com/shalb/cluster.dev/pkg/project"
	"github.com/shalb/cluster.dev/pkg/utils"
)

type CreateFileRepresentation struct {
	File    string `yaml:"file"`
	Content string `yaml:"content"`
}

// OperationConfig type that describe apply, plan and destroy operations.
type OperationConfig struct {
	Commands []interface{} `yaml:"commands" json:"commands"`
}

type OutputsConfigSpec struct {
	Command   string `yaml:"command,omitempty" json:"command,omitempty"`
	Type      string `yaml:"type" json:"type"`
	Regexp    string `yaml:"regexp,omitempty" json:"regexp,omitempty"`
	Separator string `yaml:"separator,omitempty" json:"separator,omitempty"`
}

type StateConfigFileSpec struct {
	Mask      string `yaml:"mask"`
	Dir       string `yaml:"dir"`
	Recursive bool   `yaml:"recursive"`
}

type StateConfigSpec struct {
	SaveFiles    []StateConfigFileSpec `yaml:"files"`
	SaveApplyCmd bool                  `yaml:"apply_cmd"`
	SaveEnv      bool                  `yaml:"env"`
}

type outputParser func(string, interface{}) error

// Unit describe cluster.dev shell module.
type Unit struct {
	statePtr       *StateSpec
	stackPtr       *project.Stack
	projectPtr     *project.Project
	backendPtr     project.Backend
	dependencies   []*project.DependencyOutput
	outputs        map[string]*project.DependencyOutput
	filesList      map[string][]byte
	specRaw        map[string]interface{}
	markers        map[string]interface{}
	outputRaw      []byte
	cacheDir       string
	MyName         string                     `yaml:"name"`
	WorkDir        string                     `yaml:"work_dir,omitempty"`
	Env            interface{}                `yaml:"env,omitempty"`
	CreateFiles    []CreateFileRepresentation `yaml:"create_files,omitempty"`
	ApplyConf      OperationConfig            `yaml:"apply"`
	PlanConf       OperationConfig            `yaml:"plan,omitempty"`
	DestroyConf    OperationConfig            `yaml:"destroy"`
	GetOutputsConf OutputsConfigSpec          `yaml:"outputs,omitempty"`
	StateConf      StateConfigSpec            `yaml:"state,omitempty"`
	outputParsers  map[string]outputParser
	applied        bool
}

// WasApplied return true if unit's method Apply was runned.
func (m *Unit) WasApplied() bool {
	return m.applied
}

// JSONOutputParser parse in (expected JSON string)
// and stores it in the value pointed to by out.
func (m *Unit) JSONOutputParser(in string, out interface{}) error {
	return utils.JSONDecode([]byte(in), out)
}

// RegexOutputParser parse each line od in string with key/value regexp
// and stores result as a map in the value pointed to by out.
func (m *Unit) RegexOutputParser(in string, out interface{}) error {
	rv := reflect.ValueOf(out)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return fmt.Errorf("can't set unaddressable value")
	}
	lines := strings.Split(in, "\n")
	if len(lines) == 0 {
		return nil
	}

	outTmp := make(map[string]interface{})
	for _, ln := range lines {
		if len(ln) == 0 {
			// ignore empty string
			continue
		}
		re, err := regexp.Compile(m.GetOutputsConf.Regexp)
		if err != nil {
			return err
		}
		parsed := re.FindStringSubmatch(ln)
		log.Warnf("Regexp: %v %q", m.GetOutputsConf.Regexp, re)
		if len(parsed) < 2 {
			// ignore "not found" and show warn
			log.Warnf("can't parse the output string '%v' with regexp '%v'", ln, m.GetOutputsConf.Regexp)
			continue
		}
		// Use first occurrence as key and value.
		outTmp[parsed[1]] = parsed[2]
	}
	rv.Elem().Set(reflect.ValueOf(outTmp))
	return nil
}

// SeparatorOutputParser split each line of in string with using separator
// and stores result as a map in the value pointed to by out.
func (m *Unit) SeparatorOutputParser(in string, out interface{}) error {

	rv := reflect.ValueOf(out)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return fmt.Errorf("can't set unadresseble value")
	}
	lines := strings.Split(in, "\n")
	if len(lines) == 0 {
		return nil
	}
	outTmp := make(map[string]interface{})
	for _, ln := range lines {
		if len(ln) == 0 {
			// ignore empty string
			continue
		}
		kv := strings.SplitN(ln, m.GetOutputsConf.Separator, 2)
		if len(kv) != 2 || len(ln) < len(m.GetOutputsConf.Separator) {
			// ignore line if separator does not found
			log.Warnf("can't parse the output string '%v' , separator '%v' does not found", ln, m.GetOutputsConf.Separator)
			continue
		}
		outTmp[kv[0]] = kv[1]
	}
	rv.Elem().Set(reflect.ValueOf(outTmp))
	return nil
}

func (m *Unit) Markers() map[string]interface{} {
	return m.markers
}

func (m *Unit) FilesList() map[string][]byte {
	return m.filesList
}

func (m *Unit) ReadConfig(spec map[string]interface{}, stack *project.Stack) error {
	if stack == nil {
		return fmt.Errorf("read shell module: empty stack or project")
	}
	m.stackPtr = stack
	m.projectPtr = stack.ProjectPtr
	m.specRaw = spec
	m.backendPtr = stack.Backend
	err := utils.YAMLInterfaceToType(spec, m)
	if err != nil {
		return err
	}
	if utils.IsLocalPath(m.WorkDir) {
		m.WorkDir = filepath.Join(config.Global.WorkingDir, m.stackPtr.TemplateDir, m.WorkDir)
	}

	isDir, err := utils.CheckDir(m.WorkDir)
	if err != nil {
		return fmt.Errorf("read module '%v': check working dir: %v", m.Name(), err.Error())
	}
	if !isDir {
		return fmt.Errorf("read module: check working dir: '%v' is not a directory", m.WorkDir)
	}
	m.cacheDir = filepath.Join(m.ProjectPtr().CodeCacheDir, m.Key())
	_, ok := m.outputParsers[m.GetOutputsConf.Type]
	if !ok {
		return fmt.Errorf("read module: outputs config: unknown type '%v'", m.GetOutputsConf.Type)
	}
	return nil
}

func (m *Unit) ExpectedOutputs() map[string]*project.DependencyOutput {
	if m.outputs == nil {
		m.outputs = make(map[string]*project.DependencyOutput)
	}
	return m.outputs
}

// Name return module name.
func (m *Unit) Name() string {
	return m.MyName
}

// StackPtr return ptr to module stack.
func (m *Unit) StackPtr() *project.Stack {
	return m.stackPtr
}

// ApplyOutput return output of module applying.
func (m *Unit) ApplyOutput() []byte {
	return m.outputRaw
}

// ProjectPtr return ptr to module project.
func (m *Unit) ProjectPtr() *project.Project {
	return m.projectPtr
}

// StackName return module stack name.
func (m *Unit) StackName() string {
	return m.stackPtr.Name
}

// Backend return module backend.
func (m *Unit) Backend() project.Backend {
	return m.stackPtr.Backend
}

// Dependencies return slice of module dependencies.
func (m *Unit) Dependencies() *[]*project.DependencyOutput {
	return &m.dependencies
}

func (m *Unit) Init() error {
	return nil
}

// Apply run unit apply procedure.
func (m *Unit) Apply() error {

	// Get and remember module state before markers replacement.
	m.statePtr = m.buildState()

	err := m.ReplaceMarkers()
	if err != nil {
		return err
	}
	rn, err := executor.NewExecutor(m.cacheDir)
	if err != nil {
		log.Debug(err.Error())
		return err
	}
	for key, val := range m.Env.(map[string]interface{}) {
		rn.Env = append(rn.Env, fmt.Sprintf("%v=%v", key, val))
	}

	rn.LogLabels = []string{
		m.StackName(),
		m.Name(),
		"apply",
	}
	var errMsg []byte

	var cmd string

	for _, c := range m.ApplyConf.Commands {
		cmd += fmt.Sprintf("%v\n", c)
	}
	m.outputRaw, errMsg, err = rn.Run(cmd)
	if err != nil {
		if len(errMsg) > 1 {
			return fmt.Errorf("%v, error output:\n %v", err.Error(), string(errMsg))
		}

		return err
	}
	// Get outputs.
	if m.GetOutputsConf.Command != "" {
		cmd = m.GetOutputsConf.Command
		m.outputRaw, errMsg, err = rn.Run(cmd)
		if err != nil {
			if len(errMsg) > 1 {
				return fmt.Errorf("%v, error output:\n %v", err.Error(), string(errMsg))
			}
			return err
		}
	}
	var pOutputs map[string]interface{}
	err = m.outputParsers[m.GetOutputsConf.Type](string(m.outputRaw), &pOutputs)
	if err != nil {
		return fmt.Errorf("parse outputs '%v': %v", m.GetOutputsConf.Type, err.Error())
	}
	for _, eo := range m.outputs {
		op, exists := pOutputs[eo.Output]
		if !exists {
			return fmt.Errorf("parse outputs: unit has no output named '%v', expected by another unit", eo.Output)
		}
		eo.OutputData = op
	}
	//log.Warnf("Output: %v\nParsed:%v", string(m.outputRaw), pOutputs)
	if err == nil {
		m.applied = true
		log.Warnf("set applied to true: %v, err: %v", m.Key(), err)
	}
	return err
}

// Plan module.
func (m *Unit) Plan() error {
	return nil
}

// Destroy module.
func (m *Unit) Destroy() error {
	return nil
}

// Key return uniq module index (string key for maps).
func (m *Unit) Key() string {
	return fmt.Sprintf("%v.%v", m.StackName(), m.MyName)
}

// CodeDir return path to module code directory.
func (m *Unit) CodeDir() string {
	return m.WorkDir
}

// UpdateProjectRuntimeData update project runtime dataset, adds module outputs.
// TODO: get module outputs and write to project runtime dataset. Now this function is only for printer's module interface.
func (m *Unit) UpdateProjectRuntimeData(p *project.Project) error {
	return nil
}

// ReplaceMarkers replace all templated markers with values.
func (m *Unit) ReplaceMarkers() error {
	err := project.ScanMarkers(m.Env, project.OutputsScanner, m)
	if err != nil {
		return err
	}
	err = project.ScanMarkers(m.ApplyConf.Commands, project.OutputsScanner, m)
	if err != nil {
		return err
	}
	return nil
}

func (m *Unit) KindKey() string {
	return "shell"
}
