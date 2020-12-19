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

// Modules that can be running in parallel.
type ModulesPack []*Module

// Project describes main config with user-defined variables.
type Project struct {
	Modules           map[string]*Module
	DependencyMarkers map[string]*DependencyMarker
	InsertYAMLMarkers map[string]interface{}
	Infrastructures   map[string]*infrastructure
	TmplFunctionsMap  template.FuncMap
	Backends          map[string]Backend
	DeploySequence    []ModulesPack
}

// DependencyMarker - marker for template function AddDepMarker. Represent module dependency (remote state).
type DependencyMarker struct {
	InfraName  string
	ModuleName string
	Output     string
}

// NewProject creates init and check new infrastructure project.
func NewProject(configs [][]byte) (*Project, error) {

	project := &Project{
		DependencyMarkers: map[string]*DependencyMarker{},
		InsertYAMLMarkers: map[string]interface{}{},
		Infrastructures:   map[string]*infrastructure{},
		Modules:           map[string]*Module{},
		Backends:          map[string]Backend{},
	}

	fMap := template.FuncMap{
		"remoteState":          project.AddDepMarker,
		"insertYAML":           project.AddYAMLBlockMarker,
		"ReconcilerVersionTag": printVersion,
	}

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
	err := project.appendModules()
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
		if ok := checkDependenciesRecursive(*mod, errDepth); !ok {
			return fmt.Errorf("Unresolved dependency in module %v.%v", mod.InfraPtr.Name, mod.Name)
		}
	}
	return nil
}

func checkDependenciesRecursive(mod Module, maxDepth int) bool {
	if maxDepth == 0 {
		return false
	}
	log.Debugf("Mod: %v, depth: %v\n%+v", mod.Name, maxDepth, mod.Dependencies)
	for _, dep := range mod.Dependencies {
		if ok := checkDependenciesRecursive(*dep.Module, maxDepth-1); !ok {
			return false
		}
	}
	return true
}

func (p *Project) buildDeploySequence() error {

	modDone := map[string]*Module{}
	modWait := map[string]*Module{}

	for _, mod := range p.Modules {
		modWait[fmt.Sprintf("%s.%s", mod.InfraPtr.Name, mod.Name)] = mod
	}
	res := []ModulesPack{}
	for c := 1; c < 20; c++ {
		doneLen := len(modDone)
		modPack := ModulesPack{}
		modIterDone := map[string]*Module{}
		for _, mod := range modWait {
			modIndex := fmt.Sprintf("%s.%s", mod.InfraPtr.Name, mod.Name)
			if len(mod.Dependencies) == 0 {

				log.Infof("Mod '%s' done (%d)", modIndex, c)
				modIterDone[modIndex] = mod
				delete(modWait, modIndex)
				modPack = append(modPack, mod)
				continue
			}
			var allDepsDone bool = true
			for _, dep := range mod.Dependencies {
				if findModule(*dep.Module, modDone) == nil {
					allDepsDone = false
					break
				}
			}
			if allDepsDone {
				log.Infof("Mod '%s' with deps %v done (%d)", modIndex, mod.Dependencies, c)
				modIterDone[modIndex] = mod
				delete(modWait, modIndex)
				modPack = append(modPack, mod)
			}
		}
		for k, v := range modIterDone {
			modDone[k] = v
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

// AddDepMarker function for template. Add hash marker, witch will be replaced with desired remote state.
func (p *Project) AddDepMarker(path string) (string, error) {
	splittedPath := strings.Split(path, ".")
	if len(splittedPath) != 3 {
		return "", fmt.Errorf("bad dependency path")
	}
	dep := DependencyMarker{
		InfraName:  splittedPath[0],
		ModuleName: splittedPath[1],
		Output:     splittedPath[2],
	}
	marker := createMarker("DEP")
	p.DependencyMarkers[marker] = &dep

	return fmt.Sprintf("%s", marker), nil
}

// AddYAMLBlockMarker function for template. Add hash marker, witch will be replaced with desired block.
func (p *Project) AddYAMLBlockMarker(data interface{}) (string, error) {
	marker := createMarker("YAML")
	p.InsertYAMLMarkers[marker] = data
	return fmt.Sprintf("%s", marker), nil
}

func (p *Project) appendModules() error {
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
			return fmt.Errorf("Incompatible struct")
		}
		modulesSlice := modulesSliceIf.([]interface{})
		for _, moduleData := range modulesSlice {
			mName, ok := moduleData.(map[string]interface{})["name"]
			if !ok {
				return fmt.Errorf("Incorrect module name")
			}
			mType, ok := moduleData.(map[string]interface{})["type"]
			if !ok {
				return fmt.Errorf("Incorrect module type")
			}
			mSource, ok := moduleData.(map[string]interface{})["source"]
			if !ok {
				return fmt.Errorf("Incorrect module source")
			}
			mInputs, ok := moduleData.(map[string]interface{})["inputs"]
			if !ok {
				return fmt.Errorf("Incorrect module inputs")
			}

			bPtr, exists := p.Backends[infra.BackendName]
			if !exists {
				return fmt.Errorf("Backend '%s' not found, infra: '%s'", infra.BackendName, infra.Name)
			}
			modDeps := []*Dependency{}
			dependsOn, ok := moduleData.(map[string]interface{})["depends_on"]
			if ok {
				splDep := strings.Split(dependsOn.(string), ".")
				if len(splDep) != 2 {
					return fmt.Errorf("Incorrect module dependency '%c'", dependsOn)
				}
				infNm := splDep[0]
				if infNm == "this" {
					infNm = infra.Name
				}
				modDeps = append(modDeps, &Dependency{
					InfraName:  infNm,
					ModuleName: splDep[1],
				})
			}
			mod := Module{
				InfraPtr:     infra,
				ProjectPtr:   p,
				BackendPtr:   bPtr,
				Name:         mName.(string),
				Type:         mType.(string),
				Source:       mSource.(string),
				Dependencies: modDeps,
				Inputs:       map[string]interface{}{},
				Outputs:      map[string]bool{},
			}
			inputs := mInputs.(map[string]interface{})
			//log.Debugf("%+v", p.Modules)
			for key, val := range inputs {
				mod.Inputs[key] = val
			}
			modKey := fmt.Sprintf("%s.%s", infraName, mName)
			p.Modules[modKey] = &mod
		}
	}
	// After reads all modules to project - process templated markers and set all dependencies between modules.
	for _, mod := range p.Modules {
		err := mod.ReplaceMarkers()
		if err != nil {
			return err
		}
		modStringCheck := fmt.Sprintf("%+v", mod)
		for _, dep := range mod.Dependencies {
			if dep.Module == nil {
				if dep.ModuleName == "" || dep.InfraName == "" {
					log.Fatalf("Empty dependency in module '%v.%v'", mod.InfraPtr.Name, mod.Name)
				}
				depMod, exists := p.Modules[fmt.Sprintf("%v.%v", dep.InfraName, dep.ModuleName)]
				if !exists {
					log.Fatalf("Error in module '%v.%v' dependency, target '%v.%v' does not exist", mod.InfraPtr.Name, mod.Name, dep.InfraName, dep.ModuleName)
				}
				dep.InfraName = ""
				dep.ModuleName = ""
				dep.Module = depMod
				dep.Output = ""
			}
		}
		//log.Debugf("Mod deps: %+v", mod.Dependencies)
		for marker := range p.InsertYAMLMarkers {
			if strings.Contains(modStringCheck, marker) {
				log.Fatalf("Unprocessed yaml block pointer found in module '%s.%s', template function insertYAML can only be used as a yaml value for module inputs.", mod.InfraPtr.Name, mod.Name)
			}
		}
	}

	log.Debug("Check modules dependencies...")
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
	for mName, module := range p.Modules {
		modDir := filepath.Join(codeDir, mName)
		log.Debugf("Processing module '%v' directory: '%v'", mName, modDir)
		err := os.Mkdir(modDir, 0755)
		if err != nil {
			return err
		}
		// Create main.tf
		tfFile := filepath.Join(modDir, "main.tf")
		log.Debugf(" file: '%v'", tfFile)
		codeBlock, err := module.GenMainCodeBlockHCL()
		if err != nil {
			log.Fatal(err.Error())
			return err
		}
		if p.checkContainsMarkers(string(codeBlock)) {
			log.Fatalf("Unprocessed remote state pointer found in module '%s.%s' (module block). Template function remoteState can only be used as a yaml value or a part of yaml value.", module.InfraPtr.Name, module.Name)
		}
		ioutil.WriteFile(tfFile, codeBlock, os.ModePerm)
		if err != nil {
			return err
		}
		// Create init.tf
		tfFile = filepath.Join(modDir, "init.tf")
		log.Debugf(" file: '%v'", tfFile)
		codeBlock, err = module.GenBackendCodeBlock()
		if err != nil {
			return err
		}
		if p.checkContainsMarkers(string(codeBlock)) {
			log.Fatalf("Unprocessed remote state pointer found in module '%s.%s' (backend block). Template function remoteState can only be used as a yaml value or a part of yaml value.", module.InfraPtr.Name, module.Name)
		}
		ioutil.WriteFile(tfFile, codeBlock, os.ModePerm)
		if err != nil {
			return err
		}
		// Create remote_state.tf
		codeBlock, err = module.GenDepsRemoteStates()
		if err != nil {
			return err
		}
		if len(codeBlock) > 1 {
			tfFile = filepath.Join(modDir, "remote_state.tf")
			if p.checkContainsMarkers(string(codeBlock)) {
				log.Fatalf("Unprocessed remote state pointer found in module '%s.%s' (remote states block). Template function remoteState can only be used as a yaml value or a part of yaml value.", module.InfraPtr.Name, module.Name)
			}
			log.Debugf(" file: '%v'", tfFile)
			ioutil.WriteFile(tfFile, codeBlock, os.ModePerm)
			if err != nil {
				return err
			}
		}
		// Create outputs.tf
		codeBlock, err = module.GenOutputsBlock()
		if err != nil {
			return err
		}
		if len(codeBlock) > 1 {
			tfFile = filepath.Join(modDir, "outputs.tf")
			if p.checkContainsMarkers(string(codeBlock)) {
				log.Fatalf("Unprocessed remote state pointer found in module '%s.%s' (output block). Template function remoteState can only be used as a yaml value or a part of yaml value.", module.InfraPtr.Name, module.Name)
			}
			log.Debugf(" file: '%v'", tfFile)
			ioutil.WriteFile(tfFile, codeBlock, os.ModePerm)
			if err != nil {
				return err
			}
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
			scr, err := mod.generateScripts("apply")
			if err != nil {
				return "", err
			}
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
			scr, err := mod.generateScripts("destroy")
			if err != nil {
				return "", err
			}
			applyScript += fmt.Sprintf("# Parallel index %d", index)
			applyScript += scr
			index++
		}
	}
	return applyScript, nil
}

func processingMarkers(data interface{}, procFunc func(data reflect.Value) (reflect.Value, error)) error {
	out := reflect.ValueOf(data)
	if out.Kind() == reflect.Ptr && !out.IsNil() {
		out = out.Elem()
	}
	switch out.Kind() {
	case reflect.Slice:
		for i := 0; i < out.Len(); i++ {
			if out.Index(i).Elem().Kind() == reflect.String {
				val, err := procFunc(out.Index(i))
				if err != nil {
					return err
				}
				out.Index(i).Set(val)
			} else {
				err := processingMarkers(out.Index(i).Interface(), procFunc)
				if err != nil {
					return err
				}
			}
		}
	case reflect.Map:
		for _, key := range out.MapKeys() {
			if out.MapIndex(key).Elem().Kind() == reflect.String {
				val, err := procFunc(out.MapIndex(key))
				if err != nil {
					return err
				}
				out.SetMapIndex(key, val)
			} else {
				err := processingMarkers(out.MapIndex(key).Interface(), procFunc)
				if err != nil {
					return err
				}
			}
		}
	default:

	}
	return nil
}

func (p *Project) checkContainsMarkers(data string) bool {
	for marker := range p.DependencyMarkers {
		if strings.Contains(data, marker) {
			return true
		}
	}
	return false
}
