package terraform

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/apex/log"
	"github.com/shalb/cluster.dev/internal/config"
	"github.com/shalb/cluster.dev/pkg/project"
)

// remoteStateMarkerCatName - name of markers category for remote states
const remoteStateMarkerCatName = "RemoteStateMarkers"
const insertYAMLMarkerCatName = "insertYAMLMarkers"

// TFModuleDriver driver for module type 'terraform'
type TFModuleDriver struct {
	projectPtr *project.Project
}

// NewModule creates the new TerraformModule.
func (d *TFModuleDriver) NewModule(spec map[string]interface{}, infra *project.Infrastructure) (project.Module, error) {
	mName, ok := spec["name"]
	if !ok {
		return nil, fmt.Errorf("Incorrect module name")
	}
	mType, ok := spec["type"]
	if !ok {
		return nil, fmt.Errorf("Incorrect module type")
	}
	if mType != moduleTypeKey {
		return nil, fmt.Errorf("Incorrect data for module. Expected '%v', received '%v'", moduleTypeKey, mType)
	}

	mSource, ok := spec["source"]
	if !ok {
		return nil, fmt.Errorf("Incorrect module source")
	}
	mInputs, ok := spec["inputs"]
	if !ok {
		return nil, fmt.Errorf("Incorrect module inputs")
	}
	modDeps := []*project.Dependency{}
	dependsOn, ok := spec["depends_on"]
	if ok {
		splDep := strings.Split(dependsOn.(string), ".")
		if len(splDep) != 2 {
			return nil, fmt.Errorf("Incorrect module dependency '%v'", dependsOn)
		}
		infNm := splDep[0]
		if infNm == "this" {
			infNm = infra.Name
		}
		modDeps = append(modDeps, &project.Dependency{
			InfraName:  infNm,
			ModuleName: splDep[1],
		})
		log.Debugf("Dep added: %v.%v", infNm, splDep[1])
	}

	bPtr, exists := infra.ProjectPtr.Backends[infra.BackendName]
	if !exists {
		return nil, fmt.Errorf("Backend '%s' not found, infra: '%s'", infra.BackendName, infra.Name)
	}

	mod := TFModule{
		infraPtr:                infra,
		projectPtr:              infra.ProjectPtr,
		name:                    mName.(string),
		Type:                    mType.(string),
		Source:                  mSource.(string),
		dependenciesRemoteState: modDeps,
		Inputs:                  mInputs.(map[string]interface{}),
		expectedRemoteStates:    map[string]bool{},
		preHook:                 nil,
		BackendPtr:              bPtr,
	}

	modPreHook, ok := spec["pre_hook"]
	if ok {
		hook, ok := modPreHook.(map[string]interface{})
		if !ok {
			log.Fatalf("Pre_hook configuration error.")
		}
		cmd, cmdExists := hook["command"].(string)
		script, scrExists := hook["script"].(string)
		if cmdExists && scrExists {
			log.Fatalf("Error in pre_hook config, use 'script' or 'command' option, not both")
		}
		if !cmdExists && !scrExists {
			log.Fatalf("Error in pre_hook config, use one of 'script' or 'command' option")
		}
		var ScriptData []byte
		var err error
		if cmdExists {
			ScriptData = []byte(fmt.Sprintf("#!/usr/bin/env bash\nset -e\n\n%s", cmd))
		} else {
			ScriptData, err = ioutil.ReadFile(filepath.Join(config.Global.WorkingDir, script))
			if err != nil {
				log.Fatalf("can't load pre hook script: %v", err.Error())
			}
		}
		mod.preHook = ScriptData
	}
	modPostHook, ok := spec["post_hook"]
	if ok {
		hook, ok := modPostHook.(map[string]interface{})
		if !ok {
			log.Fatalf("post_hook configuration error.")
		}
		cmd, cmdExists := hook["command"].(string)
		script, scrExists := hook["script"].(string)
		if cmdExists && scrExists {
			log.Fatalf("Error in post_hook config, use 'script' or 'command' option, not both")
		}
		if !cmdExists && !scrExists {
			log.Fatalf("Error in post_hook config, use one of 'script' or 'command' option")
		}
		var ScriptData []byte
		var err error
		if cmdExists {
			ScriptData = []byte(fmt.Sprintf("#!/usr/bin/env bash\nset -e\n\n%s", cmd))
		} else {
			ScriptData, err = ioutil.ReadFile(filepath.Join(config.Global.WorkingDir, script))
			if err != nil {
				log.Fatalf("can't load post hook script: %v", err.Error())
			}
		}
		mod.postHook = ScriptData
	}
	return &mod, nil
}

// AddRemoteStateMarker function for template. Add hash marker, witch will be replaced with desired remote state.
func (d *TFModuleDriver) AddRemoteStateMarker(path string) (string, error) {

	_, ok := d.projectPtr.Markers[remoteStateMarkerCatName]
	if !ok {
		d.projectPtr.Markers[remoteStateMarkerCatName] = map[string]*project.Dependency{}
	}
	splittedPath := strings.Split(path, ".")
	if len(splittedPath) != 3 {
		return "", fmt.Errorf("bad dependency path")
	}
	dep := project.Dependency{
		Module:     nil,
		InfraName:  splittedPath[0],
		ModuleName: splittedPath[1],
		Output:     splittedPath[2],
	}
	marker := d.projectPtr.CreateMarker("remoteState")
	d.projectPtr.Markers[remoteStateMarkerCatName].(map[string]*project.Dependency)[marker] = &dep
	log.Debugf("DEP MARKER: %+v", dep)

	return fmt.Sprintf("%s", marker), nil
}

// GetTemplateFunctions return list of additional template functions for this module driver.
func (d *TFModuleDriver) GetTemplateFunctions() map[string]interface{} {
	return map[string]interface{}{
		"remoteState": d.AddRemoteStateMarker,
		"insertYAML":  d.addYAMLBlockMarker,
	}
}

// GetScanners return list of marker scanners for this module driver.
func (d *TFModuleDriver) GetScanners() []project.MarkerScanner {
	return []project.MarkerScanner{
		remoteStatesScanner,
	}
}

// remoteStatesScanner - project scanner function, witch process dependencies markers in module data setted by AddRemoteStateMarker template function.
func remoteStatesScanner(data reflect.Value, module project.Module) (reflect.Value, error) {
	subVal := reflect.ValueOf(data.Interface())
	resString := subVal.String()
	depMarkers, ok := module.ProjectPtr().Markers[remoteStateMarkerCatName]
	if !ok {
		return subVal, nil
	}
	for key, marker := range depMarkers.(map[string]*project.Dependency) {
		if strings.Contains(resString, key) {
			if marker.InfraName == "this" {
				marker.InfraName = module.InfraName()
			}
			modKey := fmt.Sprintf("%s.%s", marker.InfraName, marker.ModuleName)
			depModule, exists := module.ProjectPtr().Modules[modKey]
			if !exists {
				log.Fatalf("Depend module does not exists. Src: '%s.%s', depend: '%s'", module.InfraName(), module.Name(), modKey)
			}
			thisModTf, ok := module.Self().(*TFModule)
			if ok {
				thisModTf.dependenciesRemoteState = append(thisModTf.dependenciesRemoteState, marker)
			} else {
				log.Fatal("Internal error")
			}

			modTf, ok := depModule.Self().(*TFModule)
			if ok {
				modTf.expectedRemoteStates[marker.Output] = true
			} else {
				log.Debugf("Ignore adding outputs to module '%v', unknown type.", module.Name())
				log.Fatal("Internal error")
			}
			return reflect.ValueOf(resString), nil
		}
	}
	return subVal, nil
}
func yamlBlockMarkerScanner(data reflect.Value, module project.Module) (reflect.Value, error) {
	subVal := reflect.ValueOf(data.Interface())

	yamlMarkers, ok := module.ProjectPtr().Markers[insertYAMLMarkerCatName].(map[string]interface{})
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

// addYAMLBlockMarker function for template. Add hash marker, witch will be replaced with desired block.
func (d *TFModuleDriver) addYAMLBlockMarker(data interface{}) (string, error) {
	markers := map[string]interface{}{}
	marker := d.projectPtr.CreateMarker("YAML")
	markers[marker] = data
	d.projectPtr.Markers[insertYAMLMarkerCatName] = markers
	return fmt.Sprintf("%s", marker), nil
}
