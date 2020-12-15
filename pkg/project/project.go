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
	"gopkg.in/yaml.v3"
)

// Project describes main config with user-defined variables.
type Project struct {
	Modules           map[string]*Module
	DependencyMarkers map[string]DependencyMarker
	InsertYAMLMarkers map[string]interface{}
	Infrastructures   map[string]*infrastructure
	TmplFunctionsMap  template.FuncMap
	Backends          map[string]Backend
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
		DependencyMarkers: map[string]DependencyMarker{},
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

func checkDependenciesRecursive(mod Module, errDepth int) bool {
	if errDepth == 0 {
		return false
	}
	// log.Debugf("Mod: %v, depth: %v", mod.Name, errDepth)
	for _, dep := range mod.Dependencies {
		if ok := checkDependenciesRecursive(*dep.Module, errDepth-1); !ok {
			return false
		}
	}
	return true
}

// func (p *Project) checkGraph() error {

// 	modDone := map[string]*Module{}
// 	modWait := map[string]*Module{}

// 	for _, mod := range p.Modules {
// 		modWait[fmt.Sprintf("%s.%s", mod.InfraPtr.Name, mod.Name)] = mod
// 	}
// 	for c := 1; c < 20; c++ {
// 		doneLen := len(modDone)
// 		for _, mod := range modWait {
// 			modIndex := fmt.Sprintf("%s.%s", mod.InfraPtr.Name, mod.Name)
// 			if len(mod.Dependencies) == 0 {

// 				log.Infof("Mod '%s' done (%d)", modIndex, c)
// 				modDone[modIndex] = mod
// 				delete(modWait, modIndex)
// 				continue
// 			}
// 			var allDepsDone bool = true
// 			for _, dep := range mod.Dependencies {
// 				if findModule(dep.Infra, dep.Module, modDone) == nil {
// 					allDepsDone = false
// 					break
// 				}
// 			}
// 			if allDepsDone {
// 				log.Infof("Mod '%s' with deps %v done (%d)", modIndex, mod.Dependencies, c)
// 				modDone[modIndex] = mod
// 				delete(modWait, modIndex)
// 			}
// 		}
// 		if doneLen == len(modDone) {
// 			log.Fatalf("Unresolved dependency %v", modWait)
// 			return fmt.Errorf("Unresolved dependency %v", modWait)
// 		}
// 		if len(modWait) == 0 {
// 			return nil
// 		}
// 	}
// 	return nil
// }

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
	p.DependencyMarkers[marker] = dep

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
			mod := Module{
				InfraPtr:     infra,
				ProjectPtr:   p,
				BackendPtr:   bPtr,
				Name:         mName.(string),
				Type:         mType.(string),
				Source:       mSource.(string),
				Inputs:       map[string]interface{}{},
				Dependencies: []Dependency{},
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
		for marker := range p.DependencyMarkers {
			if strings.Contains(modStringCheck, marker) {
				log.Fatalf("Unprocessed remote state pointer found in module '%s.%s', template function remoteState can only be used as a yaml value or a part of yaml value.", mod.InfraPtr.Name, mod.Name)
			}
		}
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
	return nil
}

// GenCode generate all terraform code for project.
func (p *Project) GenCode(codeStructName string) error {
	baseOutDir := filepath.Join("./", ".outputs")
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

		tfFile := filepath.Join(modDir, "main.tf")

		log.Debugf(" file: '%v'", tfFile)
		codeBlock, err := module.GenMainCodeBlockHCL()
		if err != nil {
			log.Fatal(err.Error())
			return err
		}
		ioutil.WriteFile(tfFile, codeBlock, os.ModePerm)

		tfFile = filepath.Join(modDir, "init.tf")
		log.Debugf(" file: '%v'", tfFile)
		codeBlock, err = module.GenBackendCodeBlock()
		if err != nil {
			return err
		}

		ioutil.WriteFile(tfFile, codeBlock, os.ModePerm)
		codeBlock, err = module.GenDepsRemoteStates()
		if err != nil {
			return err
		}
		if len(codeBlock) > 1 {
			tfFile = filepath.Join(modDir, "remote_state.tf")
			log.Debugf(" file: '%v'", tfFile)
			ioutil.WriteFile(tfFile, codeBlock, os.ModePerm)
		}
	}
	return nil
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
