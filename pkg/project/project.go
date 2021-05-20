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
type MarkerScanner func(data reflect.Value, module Module) (reflect.Value, error)

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
	ModulesOutputs  map[string]interface{}
	PrintersOutputs []PrinterOutput
}

// Project describes main config with user-defined variables.
type Project struct {
	name             string
	Modules          map[string]Module
	Infrastructures  map[string]*Infrastructure
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
}

// NewEmptyProject creates new empty project. The configuration will not be loaded.
func NewEmptyProject() *Project {
	project := &Project{
		Infrastructures:  make(map[string]*Infrastructure),
		Modules:          make(map[string]Module),
		Backends:         make(map[string]Backend),
		Markers:          make(map[string]interface{}),
		TmplFunctionsMap: templateFunctionsMap,
		objects:          make(map[string][]ObjectData),
		configData:       make(map[string]interface{}),
		secrets:          make(map[string]Secret),
		RuntimeDataset: RuntimeData{
			ModulesOutputs:  make(map[string]interface{}),
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
// Infrastructures, backends and other objects are not loads.
func LoadProjectBase() (*Project, error) {

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
		log.Fatalf("Loading project: parsing project config: ", err.Error())
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
	err = project.readModules()
	if err != nil {
		return nil, err
	}
	err = project.prepareModules()
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
	err = p.readInfrastructures()
	if err != nil {
		return err
	}
	return nil
}

func (p *Project) checkGraph() error {
	errDepth := 15
	for _, mod := range p.Modules {
		if ok := checkDependenciesRecursive(mod, errDepth); !ok {
			return fmt.Errorf("Unresolved dependency in module %v.%v", mod.InfraName(), mod.Name())
		}
	}
	return nil
}

func (p *Project) readModules() error {
	// Read modules from all infrastructures.
	for infraName, infra := range p.Infrastructures {
		for _, infraTmpl := range infra.Templates {
			for _, moduleData := range infraTmpl.Modules {
				mod, err := NewModule(moduleData, infra)
				if err != nil {
					traceModuleView, errYaml := yaml.Marshal(moduleData)
					if errYaml != nil {
						traceModuleView = []byte{}
					}
					return fmt.Errorf("infra '%v', reading modules: %v\nModule data:\n%v", infraName, err.Error(), string(traceModuleView))
				}
				if _, exists := p.Modules[mod.Key()]; exists {
					return fmt.Errorf("infra '%v', reading modules: duplicate module name: %v", infraName, mod.Name())
				}
				p.Modules[mod.Key()] = mod
			}
		}
	}
	return nil
}

func (p *Project) prepareModules() error {
	// After reads all modules to project - process templated markers and set all dependencies between modules.
	for _, mod := range p.Modules {
		err := mod.ReplaceMarkers()
		if err != nil {
			return err
		}
		if err = BuildModuleDeps(mod); err != nil {
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
	table.SetHeader([]string{"Name", "Infra count", "Modules count", "Backends count", "Secrets count"})
	table.Append([]string{
		p.name,
		fmt.Sprintf("%v", len(p.Infrastructures)),
		fmt.Sprintf("%v", len(p.Modules)),
		fmt.Sprintf("%v", len(p.Backends)),
		fmt.Sprintf("%v", len(p.secrets)),
	})
	table.Render()

	fmt.Println("Infrastructures:")
	table = tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Name", "Modules count", "Backend name", "Backend type"})
	for name, infra := range p.Infrastructures {
		mCount := 0
		for _, mod := range p.Modules {
			if mod.InfraName() == infra.Name {
				mCount++
			}
		}
		table.Append([]string{
			name,
			fmt.Sprintf("%v", mCount),
			infra.Backend.Name(),
			infra.Backend.Provider(),
		})
	}
	table.Render()

	fmt.Println("Modules:")
	table = tablewriter.NewWriter(os.Stdout)
	table.SetRowLine(true)
	// table.SetRowSeparator(".")
	table.SetHeader([]string{"Name", "Infra", "Kind", "Dependencies"})
	for name, mod := range p.Modules {
		deps := ""
		for i, dep := range *mod.Dependencies() {
			deps = fmt.Sprintf("%s%s.%s", deps, dep.InfraName, dep.ModuleName)
			if dep.Output != "" {
				deps = fmt.Sprintf("%s.%s", deps, dep.Output)
			}
			if i != len(*mod.Dependencies())-1 {
				deps += "\n"
			}
		}
		table.Append([]string{
			name,
			mod.InfraName(),
			mod.KindKey(),
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
		log.Infof("Printer: '%v', Output:\n%v", o.Name, color.Style{color.FgGreen, color.OpBold}.Sprintf(o.Output))
	}
	return nil
}
