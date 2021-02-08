package project

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"text/template"

	"github.com/apex/log"
	"github.com/shalb/cluster.dev/pkg/config"
	"gopkg.in/yaml.v3"
)

// ConfigFileName name of required project config file.
const ConfigFileName = "project.yaml"

// ModulesPack that can be running in parallel.
type ModulesPack []Module

// MarkerScanner type witch describe function for scaning markers in templated and unmarshaled yaml data.
type MarkerScanner func(data reflect.Value, module Module) (reflect.Value, error)

// Project describes main config with user-defined variables.
type Project struct {
	name             string
	Modules          map[string]Module
	Infrastructures  map[string]*Infrastructure
	TmplFunctionsMap template.FuncMap
	Backends         map[string]Backend
	Markers          map[string]interface{}
	objects          map[string][]interface{}
	config           map[string]interface{}
}

// NewProject creates init and check new project.
func NewProject(projectConf []byte, configs [][]byte) (*Project, error) {

	if projectConf == nil {
		log.Fatalf("Error reading project configuration file '%v', empty config or file does not exists.", ConfigFileName)
	}

	project := &Project{
		Infrastructures:  map[string]*Infrastructure{},
		Modules:          map[string]Module{},
		Backends:         map[string]Backend{},
		Markers:          map[string]interface{}{},
		TmplFunctionsMap: templateFunctionsMap,
		objects:          map[string][]interface{}{},
		config:           map[string]interface{}{},
	}
	for _, drv := range TemplateDriversMap {
		drv.AddTemplateFunctions(project)
	}
	var prjConfParsed map[string]interface{}
	err := yaml.Unmarshal(projectConf, &prjConfParsed)
	if err != nil {
		log.Fatal(err.Error())
	}
	if name, ok := prjConfParsed["name"].(string); !ok {
		log.Fatal("Error in project config. Name is required.")
	} else {
		project.name = name
	}

	if kn, ok := prjConfParsed["kind"].(string); !ok || kn != "project" {
		log.Fatal("Error in project config. Kind is required.")
	}

	project.config["project"] = prjConfParsed

	for _, cnf := range configs {
		tmpl, err := template.New("main").Funcs(project.TmplFunctionsMap).Option("missingkey=default").Parse(string(cnf))

		if err != nil {
			return nil, err
		}

		templatedConf := bytes.Buffer{}
		err = tmpl.Execute(&templatedConf, project.config)
		if err != nil {
			return nil, err
		}
		project.readObjects(templatedConf.Bytes())
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

func (p *Project) readObjects(objData []byte) error {

	objs, err := ReadYAMLObjects(objData)
	if err != nil {
		return err
	}
	for _, obj := range objs {
		objKind, ok := obj.(map[string]interface{})["kind"].(string)
		if !ok {
			log.Fatal("infra must contain field 'kind'")
		}
		if _, exists := p.objects[objKind]; !exists {
			p.objects[objKind] = []interface{}{}
		}
		p.objects[objKind] = append(p.objects[objKind], obj)
	}
	return nil
}

func (p *Project) prepareObjects() error {
	// Read and parse backends.
	bks, exists := p.objects[backendObjKindKey]
	if !exists {
		err := fmt.Errorf("no backend found, at least one backend needed")
		log.Debug(err.Error())
		return err
	}
	for _, bk := range bks {
		p.readBackendObj(bk.(map[string]interface{}))
	}

	// Read and parse infrastructures.
	infras, exists := p.objects[infraObjKindKey]
	if !exists {
		err := fmt.Errorf("no infrastructures found, at least one backend needed")
		log.Debug(err.Error())
		return err
	}
	for _, infra := range infras {
		err := p.readInfrastructureObj(infra.(map[string]interface{}))
		if err != nil {
			return err
		}
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
		infrastructureTemplate := make(map[string]interface{})
		err := yaml.Unmarshal(infra.Template, &infrastructureTemplate)
		if err != nil {
			log.Debugf("Can't unmarshal infrastructure template: %v", string(infra.Template))
			return err
		}
		// log.Debugf("%+v\n", infrastructureTemplate)
		modulesSliceIf, ok := infrastructureTemplate["modules"]
		if !ok {
			return fmt.Errorf("Incorrect template in infra '%v'", infraName)
		}
		modulesSlice, ok := modulesSliceIf.([]interface{})
		if !ok {
			return fmt.Errorf("Incorrect template in infra '%v'", infraName)
		}
		for _, moduleData := range modulesSlice {
			mod, err := NewModule(moduleData.(map[string]interface{}), infra)
			if err != nil {
				log.Debugf("module '%v',%v", moduleData, err.Error())
				return err
			}
			p.Modules[mod.Key()] = mod
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

	log.Info("Check modules dependencies...")
	if err := p.checkGraph(); err != nil {
		return err
	}
	return nil
}

// Build generate all terraform code for project.
func (p *Project) Build() error {
	baseOutDir := config.Global.TmpDir
	if _, err := os.Stat(baseOutDir); os.IsNotExist(err) {
		err := os.Mkdir(baseOutDir, 0755)
		if err != nil {
			return err
		}
	}
	codeDir := filepath.Join(baseOutDir, p.name)
	log.Debugf("Creates code directory: '%v'", codeDir)
	if _, err := os.Stat(codeDir); os.IsNotExist(err) {
		err := os.Mkdir(codeDir, 0755)
		if err != nil {
			return err
		}
	}
	if !config.Global.UseCache {
		log.Debugf("Remove all old content: %s", codeDir)
		err := removeDirContent(codeDir)
		if err != nil {
			log.Debug(err.Error())
			return err
		}
	}
	for _, module := range p.Modules {
		if err := module.Build(codeDir); err != nil {
			return err
		}
	}
	return nil
}

func (p *Project) Destroy() error {

	grph := grapher{}
	grph.Init(p, 1, true)

	for {
		if grph.Len() == 0 {
			return nil
		}
		md, fn, err := grph.GetNext()
		if err != nil {
			log.Errorf("error in module %v, waiting for all running modules done.", md.Key())
			grph.Wait()
			return fmt.Errorf("error in module %v:\n%v", md.Key(), err.Error())
		}
		if md == nil {
			return nil
		}
		go func(mod Module, finFunc func(error)) {
			res := mod.Destroy()
			finFunc(res)
		}(md, fn)
	}
}

func (p *Project) Apply() error {

	grph := grapher{}
	grph.Init(p, config.Global.MaxParallel, false)

	for {
		if grph.Len() == 0 {
			return nil
		}
		md, fn, err := grph.GetNext()
		if err != nil {
			log.Errorf("error in module %v, waiting for all running modules done.", md.Key())
			grph.Wait()
			return fmt.Errorf("error in module %v:\n%v", md.Key(), err.Error())
		}
		if md == nil {
			return nil
		}
		go func(mod Module, finFunc func(error)) {
			res := mod.Apply()
			finFunc(res)
		}(md, fn)
	}
}

func (p *Project) Plan() error {

	grph := grapher{}
	grph.Init(p, 1, false)

	for {
		if grph.Len() == 0 {
			return nil
		}
		md, fn, err := grph.GetNext()
		if err != nil {
			log.Errorf("error in module %v, waiting for all running modules done.", md.Key())
			grph.Wait()
			return fmt.Errorf("error in module %v:\n%v", md.Key(), err.Error())
		}
		if md == nil {
			return nil
		}
		go func(mod Module, finFunc func(error)) {
			res := mod.Plan()
			finFunc(res)
		}(md, fn)
	}
}

// Name return project name.
func (p *Project) Name() string {
	return p.name
}
