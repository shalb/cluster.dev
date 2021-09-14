package common

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"reflect"
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
	Commands interface{} `yaml:"commands" json:"commands"`
}

type GetOutputsConfig struct {
	Command     string `yaml:"command,omitempty"`
	Type        string `yaml:"type"`
	KeyRegexp   string `yaml:"key,omitempty"`
	ValueRegexp string `yaml:"value,omitempty"`
	Separator   string `yaml:"separator,omitempty"`
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

// Module describe cluster.dev shell module.
type Module struct {
	infraPtr        *project.Infrastructure
	projectPtr      *project.Project
	backendPtr      project.Backend
	dependencies    []*project.Dependency
	expectedOutputs map[string]bool
	filesList       map[string][]byte
	specRaw         map[string]interface{}
	markers         map[string]interface{}
	outputRaw       []byte
	cacheDir        string
	MyName          string                     `yaml:"name"`
	WorkDir         string                     `yaml:"work_dir,omitempty"`
	Env             interface{}                `yaml:"env,omitempty"`
	CreateFiles     []CreateFileRepresentation `yaml:"create_files,omitempty"`
	ApplyConf       OperationConfig            `yaml:"apply"`
	PlanConf        OperationConfig            `yaml:"plan,omitempty"`
	DestroyConf     OperationConfig            `yaml:"destroy"`
	GetOutputsConf  GetOutputsConfig           `yaml:"outputs,omitempty"`
	StateConf       StateConfigSpec            `yaml:"state,omitempty"`
	outputParsers   map[string]outputParser
}

func (m *Module) JSONOutputParser(in string, out interface{}) error {
	return utils.JSONDecode([]byte(in), out)
}

func (m *Module) RegexOutputParser(in string, out interface{}) error {
	return utils.JSONDecode([]byte(in), out)
}

func (m *Module) SeparatorOutputParser(in string, out interface{}) error {
	rv := reflect.ValueOf(out)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return nil
	}
	lines := strings.Split(in, "\n")
	if len(lines) == 0 {
		return nil
	}
	outTmp := make(map[string]string)
	for _, ln := range lines {
		kv := strings.SplitN(ln, m.GetOutputsConf.Separator, 2)
		if len(kv) != 2 {
			continue
		}
		outTmp[kv[0]] = kv[1]
	}
	rv.Elem().Set(reflect.ValueOf(outTmp))
	return nil
}

func (m *Module) Markers() map[string]interface{} {
	if m.markers == nil {
		m.markers = make(map[string]interface{})
	}
	return m.markers
}

func (m *Module) FilesList() map[string][]byte {
	return m.filesList
}

func (m *Module) ReadConfig(spec map[string]interface{}, infra *project.Infrastructure) error {
	if infra == nil {
		return fmt.Errorf("read shell module: empty infra or project")
	}
	m.infraPtr = infra
	m.projectPtr = infra.ProjectPtr
	m.specRaw = spec
	m.backendPtr = infra.Backend
	err := utils.YAMLInterfaceToType(spec, m)
	if err != nil {
		return err
	}
	if utils.IsLocalPath(m.WorkDir) {
		m.WorkDir = filepath.Join(config.Global.WorkingDir, m.infraPtr.TemplateDir, m.WorkDir)
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

func (m *Module) ExpectedOutputs() map[string]bool {
	if m.expectedOutputs == nil {
		m.expectedOutputs = make(map[string]bool)
	}
	return m.expectedOutputs
}

// Name return module name.
func (m *Module) Name() string {
	return m.MyName
}

// InfraPtr return ptr to module infrastructure.
func (m *Module) InfraPtr() *project.Infrastructure {
	return m.infraPtr
}

// ApplyOutput return output of module applying.
func (m *Module) ApplyOutput() []byte {
	return m.outputRaw
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
	rn, err := executor.NewExecutor(m.cacheDir)
	if err != nil {
		log.Debug(err.Error())
		return err
	}
	for key, val := range m.Env.(map[string]interface{}) {
		rn.Env = append(rn.Env, fmt.Sprintf("%v=%v", key, val))
	}

	rn.LogLabels = []string{
		m.InfraName(),
		m.Name(),
		"apply",
	}
	var errMsg []byte

	var cmd string

	// cmd = "set -e\n"
	for _, c := range m.ApplyConf.Commands.([]interface{}) {
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
	var pOutputs interface{}
	err = m.outputParsers[m.GetOutputsConf.Type](string(m.outputRaw), &pOutputs)
	if err != nil {
		if e, ok := err.(*json.SyntaxError); ok {
			return fmt.Errorf("parse outputs: '%v': %v", m.GetOutputsConf.Type, e.Offset)
		}
		return fmt.Errorf("parse outputs '%v': %v", m.GetOutputsConf.Type, err.Error())
	}
	log.Warnf("Output: %v\nParsed:%v", string(m.outputRaw), pOutputs)
	return err
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
	return fmt.Sprintf("%v.%v", m.InfraName(), m.MyName)
}

// CodeDir return path to module code directory.
func (m *Module) CodeDir() string {
	return m.WorkDir
}

// UpdateProjectRuntimeData update project runtime dataset, adds module outputs.
// TODO: get module outputs and write to project runtime dataset. Now this function is only for printer's module interface.
func (m *Module) UpdateProjectRuntimeData(p *project.Project) error {
	return nil
}

// ReplaceMarkers replace all templated markers with values.
func (m *Module) ReplaceMarkers() error {

	tEnv := m.Env
	err := project.ScanMarkers(&tEnv, project.YamlBlockMarkerScanner, m)
	if err != nil {
		return err
	}
	err = project.ScanMarkers(&tEnv, project.OutputsScanner, m)
	if err != nil {
		return err
	}
	m.Env = tEnv

	tmpAppl := m.ApplyConf.Commands
	err = project.ScanMarkers(&tmpAppl, project.YamlBlockMarkerScanner, m)
	if err != nil {
		return err
	}
	err = project.ScanMarkers(&tmpAppl, project.OutputsScanner, m)
	if err != nil {
		return err
	}
	log.Warnf("%v", tmpAppl)
	m.ApplyConf.Commands = tmpAppl

	return nil
}

func (m *Module) KindKey() string {
	return "shell"
}
