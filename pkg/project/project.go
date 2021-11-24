package project

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"text/template"

	"github.com/apex/log"
	"github.com/gookit/color"
	"github.com/olekukonko/tablewriter"
	"github.com/shalb/cluster.dev/pkg/config"
	"github.com/shalb/cluster.dev/pkg/utils"
	"gopkg.in/yaml.v3"
)

// ConfigFileName name of required project config file.
const ConfigFileName = "project.yaml"
const projectObjKindKey = "Project"

// MarkerScanner type witch describe function for scaning markers in templated and unmarshaled yaml data.
type MarkerScanner func(data reflect.Value, unit Unit) (reflect.Value, error)

// TODO:
// // ProjectConfSpec type for project.yaml config.
// type ProjectConfSpec struct {
// 	Name      string                 `yaml:"name"`
// 	Kind      string                 `yaml:"kind"`
// 	Backend   string                 `yaml:"backend,omitempty"`
// 	Exports   map[string]interface{} `yaml:"exports"`
// 	Variables map[string]interface{} `yaml:"variables"`
// }
type PrinterOutput struct {
	Name   string `json:"name"`
	Output string `json:"output"`
}

type RuntimeData struct {
	UnitsOutputs    map[string]interface{}
	PrintersOutputs []PrinterOutput
}

// Project describes main config with user-defined variables.
type Project struct {
	name             string
	Units            map[string]Unit
	Stacks           map[string]*Stack
	TmplFunctionsMap template.FuncMap
	Backends         map[string]Backend
	Markers          map[string]interface{}
	secrets          map[string]Secret
	configData       map[string]interface{}
	configDataFile   []byte
	objects          map[string][]ObjectData
	objectsFiles     map[string][]byte
	CodeCacheDir     string
	StateMutex       sync.Mutex
	InitLock         sync.Mutex
	RuntimeDataset   RuntimeData
	StateBackendName string
	OwnState         *StateProject
}

// NewEmptyProject creates new empty project. The configuration will not be loaded.
func NewEmptyProject() *Project {
	project := &Project{
		Stacks:           make(map[string]*Stack),
		Units:            make(map[string]Unit),
		Backends:         make(map[string]Backend),
		Markers:          make(map[string]interface{}),
		TmplFunctionsMap: templateFunctionsMap,
		objects:          make(map[string][]ObjectData),
		configData:       make(map[string]interface{}),
		secrets:          make(map[string]Secret),
		RuntimeDataset: RuntimeData{
			UnitsOutputs:    make(map[string]interface{}),
			PrintersOutputs: make([]PrinterOutput, 0),
		},
		CodeCacheDir: config.Global.CacheDir,
	}
	for _, drv := range TemplateDriversMap {
		drv.AddTemplateFunctions(project)
	}
	return project
}

// LoadProjectBase read project data in current directory, create base project, and load secrets.
// stacks, backends and other objects are not loads.
func LoadProjectBase() (*Project, error) {
	for mf := range UnitFactoriesMap {
		log.Debugf("Registering unit type: %v", mf)
	}
	project := NewEmptyProject()
	err := project.MkBuildDir()
	if err != nil {
		log.Fatalf("Loading project: creating working dir: '%v'.", err.Error())
	}
	err = project.readManifests()
	if project.configDataFile == nil {
		log.Fatalf("Loading project: loading project config: file '%v', empty configuration	.", ConfigFileName)
	}

	var prjConfParsed map[string]interface{}
	err = yaml.Unmarshal(project.configDataFile, &prjConfParsed)
	if err != nil {
		log.Fatalf("Loading project: parsing project config: ", utils.ResolveYamlError(project.configDataFile, err))
	}
	if name, ok := prjConfParsed["name"].(string); !ok {
		log.Fatal("Loading project: error in project config: name is required.")
	} else {
		project.name = name
	}

	if kn, ok := prjConfParsed["kind"].(string); !ok || kn != projectObjKindKey {
		log.Fatal("Loading project: error in project config: kind is required.")
	}

	if exports, ok := prjConfParsed["exports"]; ok {
		err = project.ExportEnvs(exports)
		if err != nil {
			log.Fatalf("Loading project: %v", err.Error())
		}
	}

	if stateBackend, exists := prjConfParsed["backend"].(string); exists {
		project.StateBackendName = stateBackend
	}

	project.configData["project"] = prjConfParsed

	err = project.readSecrets()
	if err != nil {
		log.Fatalf("Loading project: %v", err.Error())
	}
	return project, nil
}

// LoadProjectFull read project data in current directory, create base project, load secrets and all project's objects.
func LoadProjectFull() (*Project, error) {
	project, err := LoadProjectBase()
	if err != nil {
		return nil, fmt.Errorf("loading project: %w", err)
	}
	for filename, cnf := range project.objectsFiles {
		templatedConf, isWarn, err := project.TemplateTry(cnf)
		if project.fileIsSecret(filename) {
			// Skip secrets, which loaded in LoadProjectBase().
			continue
		}
		if err != nil {
			if isWarn {
				rel, _ := filepath.Rel(config.Global.WorkingDir, filename)
				log.Warnf("File %v has unresolved template key: \n%v", rel, err.Error())
			} else {
				log.Fatal(err.Error())
			}
		}
		err = project.readObjects(templatedConf, filename)
		if err != nil {
			log.Fatalf("load project: %v", err.Error())
		}
	}

	err = project.prepareObjects()
	if err != nil {
		return nil, err
	}
	if !config.Global.NotLoadState {
		_, err = project.LoadState()
		if err != nil {
			return nil, fmt.Errorf("loading project: %w", err)
		}
	}
	err = project.readUnits()
	if err != nil {
		return nil, err
	}
	err = project.prepareUnits()
	if err != nil {
		return nil, err
	}
	return project, nil
}

// ObjectData simple representation of project object.
type ObjectData struct {
	filename string
	data     map[string]interface{}
}

func (p *Project) readObjects(objData []byte, filename string) error {
	// Ignore secrets.
	if p.fileIsSecret(filename) {
		return nil
	}
	objs, err := utils.ReadYAMLObjects(objData)
	if err != nil {
		return err
	}
	for _, obj := range objs {
		objKind, ok := obj["kind"].(string)
		if !ok {
			log.Fatal("object must contain field 'kind'")
		}
		if _, exists := p.objects[objKind]; !exists {
			p.objects[objKind] = []ObjectData{}
		}
		p.objects[objKind] = append(p.objects[objKind], ObjectData{
			filename: filename,
			data:     obj,
		})
	}
	return nil
}

func (p *Project) prepareObjects() error {
	err := p.readBackends()
	if err != nil {
		return err
	}
	err = p.readStacks()
	if err != nil {
		return err
	}
	return nil
}

func (p *Project) checkGraph() error {
	errDepth := 15
	for _, unit := range p.Units {
		if ok := checkDependenciesRecursive(unit, errDepth); !ok {
			return fmt.Errorf("Unresolved dependency in unit %v.%v", unit.Stack().Name, unit.Name())
		}
	}
	return nil
}

func (p *Project) readUnits() error {
	// Read units from all stacks.
	for stackName, stack := range p.Stacks {
		for _, stackTmpl := range stack.Templates {
			for _, unitData := range stackTmpl.Units {
				mod, err := NewUnit(unitData, stack)
				if err != nil {
					traceUnitView, errYaml := yaml.Marshal(unitData)
					if errYaml != nil {
						traceUnitView = []byte{}
					}
					return fmt.Errorf("stack '%v', reading units: %v\nUnit data:\n%v", stackName, err.Error(), string(traceUnitView))
				}
				if _, exists := p.Units[mod.Key()]; exists {
					return fmt.Errorf("stack '%v', reading units: duplicate unit name: %v", stackName, mod.Name())
				}
				p.Units[mod.Key()] = mod
			}
		}
	}
	return nil
}

func (p *Project) prepareUnits() error {
	// After reads all units to project - process templated markers and set all dependencies between units.
	for _, mod := range p.Units {
		err := mod.ReplaceMarkers()
		if err != nil {
			return err
		}
		if err = BuildUnitsDeps(mod); err != nil {
			return err
		}
	}
	if err := p.checkGraph(); err != nil {
		return err
	}
	//p.printState()
	return nil
}

func (p *Project) MkBuildDir() error {
	baseOutDir := config.Global.TmpDir
	if _, err := os.Stat(baseOutDir); os.IsNotExist(err) {
		err := os.Mkdir(baseOutDir, 0755)
		if err != nil {
			return err
		}
	}
	relPath, err := filepath.Rel(config.Global.WorkingDir, p.CodeCacheDir)
	if err != nil {
		return err
	}
	log.Debugf("Creates code directory: './%v'", relPath)
	if _, err := os.Stat(p.CodeCacheDir); os.IsNotExist(err) {
		err := os.Mkdir(p.CodeCacheDir, 0755)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *Project) ClearCacheDir() error {
	relPath, err := filepath.Rel(config.Global.WorkingDir, p.CodeCacheDir)
	if err != nil {
		return err
	}
	log.Debugf("Creates code directory: './%v'", relPath)
	if _, err := os.Stat(p.CodeCacheDir); os.IsNotExist(err) {
		return nil
	}
	if !config.Global.UseCache {
		log.Debugf("Removes all old content: './%s'", relPath)

		err := removeDirContent(p.CodeCacheDir)
		if err != nil {
			return err
		}
		if _, err := os.Stat(config.Global.TemplatesCacheDir); os.IsNotExist(err) {
			return nil
		}
		return removeDirContent(config.Global.TemplatesCacheDir)
	}
	return nil
}

// Name return project name.
func (p *Project) Name() string {
	return p.name
}

// Return project conf and slice of others config files.
func (p *Project) readManifests() error {
	var files []string
	var err error
	files, _ = filepath.Glob(config.Global.WorkingDir + "/*.yaml")
	objFiles := make(map[string][]byte)
	for _, file := range files {
		fileName, _ := filepath.Rel(config.Global.WorkingDir, file)
		if fileName == ConfigFileName {
			p.configDataFile, err = ioutil.ReadFile(file)
		} else {
			objFiles[file], err = ioutil.ReadFile(file)
		}
		if err != nil {
			return fmt.Errorf("reading configs %v: %v", file, err)
		}
	}
	p.objectsFiles = objFiles
	return nil
}

// PrintInfo print project info.
func (p *Project) PrintInfo() error {
	fmt.Println("Project:")
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Name", "Stacks count", "units count", "Backends count", "Secrets count"})
	table.Append([]string{
		p.name,
		fmt.Sprintf("%v", len(p.Stacks)),
		fmt.Sprintf("%v", len(p.Units)),
		fmt.Sprintf("%v", len(p.Backends)),
		fmt.Sprintf("%v", len(p.secrets)),
	})
	table.Render()

	fmt.Println("Stacks:")
	table = tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Name", "units count", "Backend name", "Backend type"})
	for name, stack := range p.Stacks {
		mCount := 0
		for _, unit := range p.Units {
			if unit.Stack().Name == stack.Name {
				mCount++
			}
		}
		table.Append([]string{
			name,
			fmt.Sprintf("%v", mCount),
			stack.Backend.Name(),
			stack.Backend.Provider(),
		})
	}
	table.Render()

	fmt.Println("units:")
	table = tablewriter.NewWriter(os.Stdout)
	table.SetRowLine(true)
	// table.SetRowSeparator(".")
	table.SetHeader([]string{"Name", "Stack", "Kind", "Dependencies"})
	for name, unit := range p.Units {
		deps := ""
		for i, dep := range *unit.Dependencies() {
			deps = fmt.Sprintf("%s%s.%s", deps, dep.StackName, dep.UnitName)
			if dep.Output != "" {
				deps = fmt.Sprintf("%s.%s", deps, dep.Output)
			}
			if i != len(*unit.Dependencies())-1 {
				deps += "\n"
			}
		}
		table.Append([]string{
			name,
			unit.Stack().Name,
			unit.KindKey(),
			deps,
		})
	}
	table.Render()
	return nil
}

func (p *Project) ExportEnvs(ex interface{}) error {
	exports, correct := ex.(map[string]interface{})
	if !correct {
		return fmt.Errorf("exports: malformed exports configuration")
	}
	for key, val := range exports {
		log.Debugf("Exports: %v", key)
		valStr := fmt.Sprintf("%v", val)
		os.Setenv(key, valStr)
	}
	return nil
}

func (p *Project) PrintOutputs() error {
	for _, o := range p.RuntimeDataset.PrintersOutputs {
		if len(o.Output) > 0 {
			log.Infof("Printer: '%v', Output:\n%v", o.Name, color.Style{color.FgGreen, color.OpBold}.Sprintf(o.Output))
		}
	}
	return nil
}
