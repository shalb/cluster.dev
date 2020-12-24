package project

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/apex/log"
	"github.com/shalb/cluster.dev/internal/config"
	"gopkg.in/yaml.v3"
)

// ModulesPack that can be running in parallel.
type ModulesPack []Module

// MarkerScanner type witch describe function for scaning markers in templated and unmarshaled yaml data.
type MarkerScanner func(data reflect.Value, module Module) (reflect.Value, error)

// Project describes main config with user-defined variables.
type Project struct {
	Modules          map[string]Module
	ModuleDrivers    map[string]ModuleDriver
	Infrastructures  map[string]*Infrastructure
	TmplFunctionsMap template.FuncMap
	MarkerScaners    []MarkerScanner
	Backends         map[string]Backend
	DeploySequence   []ModulesPack
	Markers          map[string]interface{}
}

// NewProject creates init and check new infrastructure project.
func NewProject(configs [][]byte) (*Project, error) {

	project := &Project{
		Infrastructures: map[string]*Infrastructure{},
		Modules:         map[string]Module{},
		Backends:        map[string]Backend{},
		Markers:         map[string]interface{}{},
		ModuleDrivers:   map[string]ModuleDriver{},
		MarkerScaners:   []MarkerScanner{},
	}

	fMap := template.FuncMap{
		"insertYAML":           project.AddYAMLBlockMarker,
		"ReconcilerVersionTag": printVersion,
	}

	project.MarkerScaners = append(project.MarkerScaners, project.checkYAMLBlockMarker)
	for key, drvFac := range ModuleDriverFactories {
		drv := drvFac.New(project)
		project.ModuleDrivers[key] = drv
		project.MarkerScaners = append(project.MarkerScaners, drv.GetScanners()...)
		for k, f := range drv.GetTemplateFunctions() {
			_, ok := fMap[k]
			if ok {
				log.Debugf("Template function '%v' is already exists", k)
				return nil, fmt.Errorf("Template function '%v' is already exists in map", k)
			}
			fMap[k] = f
		}

	}
	log.Debugf("Fmap %v", fMap)
	project.TmplFunctionsMap = fMap

	templatedConfigs := [][]byte{}
	for _, cnf := range configs {
		tmpl, err := template.New("main").Funcs(fMap).Option("missingkey=error").Parse(string(cnf))

		if err != nil {
			return nil, err
		}

		templatedConf := bytes.Buffer{}
		err = tmpl.Execute(&templatedConf, nil)
		if err != nil {
			return nil, err
		}
		templatedConfigs = append(templatedConfigs, templatedConf.Bytes())
		//var infrastructuresList []map[string]interface{}
		dec := yaml.NewDecoder(bytes.NewReader(templatedConf.Bytes()))
		for {
			var parsedConf = make(map[string]interface{})
			err = dec.Decode(&parsedConf)
			if err != nil {
				if err.Error() == "EOF" {
					break
				}
				log.Debugf("can't decode config to yaml: %s\n%s", err.Error(), string(cnf))
				return nil, fmt.Errorf("can't decode config to yaml: %s", err.Error())
			}
			err := project.readObject(parsedConf)
			if err != nil {
				return nil, err
			}
		}

	}
	err := project.readModules()
	if err != nil {
		return nil, err
	}
	err = project.prepareModules()
	if err != nil {
		return nil, err
	}
	return project, nil
}

func (p *Project) readObject(obj map[string]interface{}) error {
	objKind, ok := obj["kind"].(string)
	if !ok {
		log.Fatal("infra must contain field 'kind'")
	}
	switch objKind {
	case "backend":
		return p.readBackendObj(obj)
	case "infrastructure":
		return p.readInfrastructureObj(obj)
	}
	return fmt.Errorf("Unknown object kind '%s'", objKind)
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

func checkDependenciesRecursive(mod Module, maxDepth int) bool {
	if maxDepth == 0 {
		return false
	}
	// log.Debugf("Mod: %v, depth: %v\n%+v", mod.Name, maxDepth, mod.Dependencies)
	for _, dep := range *mod.Dependencies() {
		if ok := checkDependenciesRecursive(*dep.Module, maxDepth-1); !ok {
			return false
		}
	}
	return true
}

func (p *Project) buildDeploySequence() error {

	modDone := map[string]Module{}
	modWait := map[string]Module{}

	for _, mod := range p.Modules {
		modWait[fmt.Sprintf("%s.%s", mod.InfraName(), mod.Name())] = mod
	}
	res := []ModulesPack{}
	for c := 1; c < 20; c++ {
		doneLen := len(modDone)
		modPack := ModulesPack{}
		modIterDone := map[string]*Module{}
		for _, mod := range modWait {
			modIndex := fmt.Sprintf("%s.%s", mod.InfraName(), mod.Name())
			if len(*mod.Dependencies()) == 0 {
				modIterDone[modIndex] = &mod
				log.Infof(" '%s' - ok", modIndex)
				delete(modWait, modIndex)
				modPack = append(modPack, mod)
				continue
			}
			var allDepsDone bool = true
			for _, dep := range *mod.Dependencies() {
				if findModule(*dep.Module, modDone) == nil {
					allDepsDone = false
					break
				}
			}
			if allDepsDone {
				log.Infof(" '%s' - ok", modIndex)
				modIterDone[modIndex] = &mod
				delete(modWait, modIndex)
				modPack = append(modPack, mod)
			}
		}
		for k, v := range modIterDone {
			modDone[k] = *v
		}
		res = append(res, modPack)
		p.DeploySequence = res
		if doneLen == len(modDone) {
			log.Fatalf("Unresolved dependency %v", modWait)
			return fmt.Errorf("Unresolved dependency %v", modWait)
		}
		if len(modWait) == 0 {
			return nil
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
		modulesSlice := modulesSliceIf.([]interface{})
		for _, moduleData := range modulesSlice {
			mod, err := NewModule(moduleData.(map[string]interface{}), infra)
			if err != nil {
				return err
			}
			modKey := fmt.Sprintf("%s.%s", infraName, mod.Name())
			p.Modules[modKey] = mod
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
		modStringCheck := fmt.Sprintf("%+v", mod)
		for _, dep := range *mod.Dependencies() {
			if dep.Module == nil {
				if dep.ModuleName == "" || dep.InfraName == "" {
					log.Fatalf("Empty dependency in module '%v.%v'", mod.InfraName(), mod.Name())
				}
				depMod, exists := p.Modules[fmt.Sprintf("%v.%v", dep.InfraName, dep.ModuleName)]
				if !exists {
					log.Fatalf("Error in module '%v.%v' dependency, target '%v.%v' does not exist", mod.InfraName(), mod.Name(), dep.InfraName, dep.ModuleName)
				}
				dep.InfraName = ""
				dep.ModuleName = ""
				dep.Module = &depMod
				dep.Output = ""
			}
		}
		//log.Debugf("Mod deps: %+v", mod.Dependencies)
		yamlMarkers, ok := p.Markers["InsertYAMLMarkers"].(map[string]interface{})
		if !ok {
			log.Debugf("Internal error.")
			return fmt.Errorf("internal error")
		}
		for marker := range yamlMarkers {
			if strings.Contains(modStringCheck, marker) {
				log.Debugf("%+v", modStringCheck)
				log.Fatalf("Unprocessed yaml block pointer found in module '%s.%s', template function insertYAML can only be used as a yaml value for module inputs.", mod.InfraName(), mod.Name())
			}
		}
	}

	log.Info("Check modules dependencies...")
	if err := p.checkGraph(); err != nil {
		return err
	}
	p.buildDeploySequence()
	return nil
}

// GenCode generate all terraform code for project.
func (p *Project) GenCode(codeStructName string) error {
	baseOutDir := config.Global.TmpDir
	if _, err := os.Stat(baseOutDir); os.IsNotExist(err) {
		err := os.Mkdir(baseOutDir, 0755)
		if err != nil {
			return err
		}
	}
	codeDir := filepath.Join(baseOutDir, codeStructName)
	log.Debugf("Creates code directory: '%v'", codeDir)
	if _, err := os.Stat(codeDir); os.IsNotExist(err) {
		err := os.Mkdir(codeDir, 0755)
		if err != nil {
			return err
		}
	}
	log.Debugf("Remove all old content: %s", codeDir)
	err := removeDirContent(codeDir)
	if err != nil {
		return err
	}
	for _, module := range p.Modules {
		if err := module.CreateCodeDir(codeDir); err != nil {
			return err
		}

	}
	script, err := p.generateScriptApply()
	if err != nil {
		return err
	}
	scriptFile := filepath.Join(codeDir, "apply.sh")
	ioutil.WriteFile(scriptFile, []byte(script), 0777)
	if err != nil {
		return err
	}
	script, err = p.generateScriptDestroy()
	if err != nil {
		return err
	}
	scriptFile = filepath.Join(codeDir, "destroy.sh")
	ioutil.WriteFile(scriptFile, []byte(script), 0777)
	if err != nil {
		return err
	}
	return nil
}

func (p *Project) generateScriptApply() (string, error) {

	applyScript := `#!/bin/bash

mkdir -p ../.tmp/plugins/

`

	for index, modPack := range p.DeploySequence {
		for _, mod := range modPack {
			scr := mod.GetApplyShellCmd()
			applyScript += fmt.Sprintf("# Parallel index %d", index)
			applyScript += scr
		}
	}
	return applyScript, nil
}

func (p *Project) generateScriptDestroy() (string, error) {

	applyScript := `#!/bin/bash

mkdir -p ../.tmp/plugins/

`
	index := 0
	for i := len(p.DeploySequence) - 1; i >= 0; i-- {
		modPack := p.DeploySequence[i]
		for _, mod := range modPack {
			scr := mod.GetDestroyShellCmd()
			applyScript += fmt.Sprintf("# Parallel index %d", index)
			applyScript += scr
			index++
		}
	}
	return applyScript, nil
}

func (p *Project) CheckContainsMarkers(data string) bool {
	for _, markerClass := range p.Markers {
		vl := reflect.ValueOf(markerClass)
		if vl.Kind() != reflect.Map {
			log.Fatal("Internal error.")
		}
		for _, marker := range vl.MapKeys() {
			if strings.Contains(data, marker.String()) {
				return true
			}
		}
	}
	return false
}

// AddYAMLBlockMarker function for template. Add hash marker, witch will be replaced with desired block.
func (p *Project) AddYAMLBlockMarker(data interface{}) (string, error) {
	markers := map[string]interface{}{}
	marker := p.CreateMarker("YAML")
	markers[marker] = data
	p.Markers["InsertYAMLMarkers"] = markers
	return fmt.Sprintf("%s", marker), nil
}

func (p *Project) checkYAMLBlockMarker(data reflect.Value, module Module) (reflect.Value, error) {
	subVal := reflect.ValueOf(data.Interface())

	yamlMarkers, ok := p.Markers["InsertYAMLMarkers"].(map[string]interface{})
	if !ok {
		log.Fatalf("Internal error.")
	}
	for hash := range yamlMarkers {
		if subVal.String() == hash {
			return reflect.ValueOf(yamlMarkers[hash]), nil
		}
	}
	return subVal, nil
}
