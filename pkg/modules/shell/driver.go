package shell

import (
	"fmt"
	"io/ioutil"
	"reflect"
	"strings"

	"github.com/apex/log"

	"github.com/shalb/cluster.dev/pkg/project"
)

// moduleTypeKey - string representation of this module type.
const moduleTypeKey = "shell"

// TFModuleDriver driver for module type
type TFModuleDriver struct {
	projectPtr *project.Project
}

// DependencyMarker - marker for template function AddDepMarker. Represent module dependency (remote state).
type DependencyMarker struct {
	InfraName  string
	ModuleName string
	Output     string
}

// NewModule creates the new TerraformModule.
func (d *TFModuleDriver) NewModule(spec map[string]interface{}, infra *project.Infrastructure) (project.Module, error) {
	mName, ok := spec["name"]
	if !ok {
		return nil, fmt.Errorf("Incorrect module name")
	}
	mType, ok := spec["type"].(string)
	if !ok {
		return nil, fmt.Errorf("Incorrect module type")
	}
	if mType != moduleTypeKey {
		return nil, fmt.Errorf("Incorrect data for module. Expected '%v', received '%v'", moduleTypeKey, mType)
	}

	mCommand, okCmd := spec["command"].(string)
	mScript, okScript := spec["script"].(string)
	if !okCmd && !okScript {
		return nil, fmt.Errorf("one of the options 'script' or 'command' required")
	}
	if okCmd && okScript {
		return nil, fmt.Errorf("only one of the options 'script' or 'command' should be used, not both")
	}
	var ScriptData []byte
	var err error
	if okCmd {
		ScriptData = []byte(fmt.Sprintf("#!/usr/bin/env bash\nset -e\n\n%s", mCommand))
	} else {
		ScriptData, err = ioutil.ReadFile(mScript)
		if err != nil {
			return nil, err
		}
	}
	mInputs, ok := spec["inputs"].([]interface{})
	if !ok {
		mInputs = []interface{}{}
	}
	mInputsStr := []string{}

	for _, arg := range mInputs {
		str, ok := arg.(string)
		if !ok {
			log.Debug("Wrong type of shell module inputs. Expected array of strings.")
		}
		mInputsStr = append(mInputsStr, str)
	}

	modDeps := []*project.Dependency{}
	dependsOn, ok := spec["depends_on"]
	if ok {
		splDep := strings.Split(dependsOn.(string), ".")
		if len(splDep) != 2 {
			return nil, fmt.Errorf("Incorrect module dependency '%c'", dependsOn)
		}
		infNm := splDep[0]
		if infNm == "this" {
			infNm = infra.Name
		}
		modDeps = append(modDeps, &project.Dependency{
			InfraName:  infNm,
			ModuleName: splDep[1],
		})
	}
	bPtr, exists := infra.ProjectPtr.Backends[infra.BackendName]
	if !exists {
		return nil, fmt.Errorf("Backend '%s' not found, infra: '%s'", infra.BackendName, infra.Name)
	}

	mod := Module{
		infraPtr:        infra,
		projectPtr:      infra.ProjectPtr,
		name:            mName.(string),
		Type:            mType,
		dependencies:    modDeps,
		Inputs:          mInputsStr,
		expectedOutputs: map[string]bool{},
		BackendPtr:      bPtr,
		scriptData:      ScriptData,
	}

	return &mod, nil
}

// GetTemplateFunctions return list of additional template functions for this module driver.
func (d *TFModuleDriver) GetTemplateFunctions() map[string]interface{} {
	return map[string]interface{}{}
}

// GetScanners return list of marker scanners for this module driver.
func (d *TFModuleDriver) GetScanners() []project.MarkerScanner {
	return []project.MarkerScanner{}
}

// OutputsReplacer - replace output markers in shell module to env variables names.
func OutputsReplacer(data reflect.Value, module project.Module) (reflect.Value, error) {
	subVal := reflect.ValueOf(data.Interface())
	resString := subVal.String()
	outputMarkers, ok := module.ProjectPtr().Markers[project.OutputMarkerCatName]
	if !ok {
		return subVal, nil
	}
	for key, dep := range outputMarkers.(map[string]*project.Dependency) {
		if strings.Contains(resString, key) {
			outputVarName := fmt.Sprintf("%v.%v.%v", dep.InfraName, dep.ModuleName, dep.Output)
			outputVarName = project.ConvertToShellVar(outputVarName)
			resString = strings.ReplaceAll(resString, key, outputVarName)
			return reflect.ValueOf(resString), nil
		}
	}
	return subVal, nil
}
