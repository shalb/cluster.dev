package terraform

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/apex/log"
	"github.com/shalb/cluster.dev/pkg/project"
)

// TFModuleDriver driver for module type 'terraform'
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

	mod := TFModule{
		infraPtr:        infra,
		projectPtr:      infra.ProjectPtr,
		name:            mName.(string),
		Type:            mType.(string),
		Source:          mSource.(string),
		dependencies:    modDeps,
		Inputs:          mInputs.(map[string]interface{}),
		ExpectedOutputs: map[string]bool{},
		BackendPtr:      bPtr,
	}

	return &mod, nil
}

// AddRemoteStateMarker function for template. Add hash marker, witch will be replaced with desired remote state.
func (d *TFModuleDriver) AddRemoteStateMarker(path string) (string, error) {

	_, ok := d.projectPtr.Markers["DependencyMarkers"]
	if !ok {
		d.projectPtr.Markers["DependencyMarkers"] = map[string]*DependencyMarker{}
	}
	splittedPath := strings.Split(path, ".")
	if len(splittedPath) != 3 {
		return "", fmt.Errorf("bad dependency path")
	}
	dep := DependencyMarker{
		InfraName:  splittedPath[0],
		ModuleName: splittedPath[1],
		Output:     splittedPath[2],
	}
	marker := d.projectPtr.CreateMarker("DEP")
	d.projectPtr.Markers["DependencyMarkers"].(map[string]*DependencyMarker)[marker] = &dep

	return fmt.Sprintf("%s", marker), nil
}

// GetTemplateFunctions return list of additional template functions for this module driver.
func (d *TFModuleDriver) GetTemplateFunctions() map[string]interface{} {
	return map[string]interface{}{
		"remoteState": d.AddRemoteStateMarker,
	}
}

// GetScanners return list of marker scanners for this module driver.
func (d *TFModuleDriver) GetScanners() []project.MarkerScanner {
	return []project.MarkerScanner{
		d.DependMarkersScanner,
	}
}

// DependMarkersScanner - project scanner function, witch process dependencies markers in module data setted by AddRemoteStateMarker template function.
func (d *TFModuleDriver) DependMarkersScanner(data reflect.Value, module project.Module) (reflect.Value, error) {
	subVal := reflect.ValueOf(data.Interface())
	resString := subVal.String()
	depMarkers, ok := d.projectPtr.Markers["DependencyMarkers"]
	if !ok {
		log.Debug("Internal error.")
	}
	for key, marker := range depMarkers.(map[string]*DependencyMarker) {
		if strings.Contains(resString, key) {
			if marker.InfraName == "this" {
				marker.InfraName = module.InfraName()
			}
			modKey := fmt.Sprintf("%s.%s", marker.InfraName, marker.ModuleName)
			depModule, exists := d.projectPtr.Modules[modKey]
			if !exists {
				log.Debugf("Depend module does not exists. Src: '%s.%s', depend: '%s'\n%+v", module.InfraName(), module.Name(), modKey, d.projectPtr.Modules)
				return reflect.ValueOf(nil), fmt.Errorf("Depend module does not exists. Src: '%s.%s', depend: '%s'", module.InfraName(), module.Name(), modKey)
			}
			deps := module.Dependencies()
			*deps = append(*deps, &project.Dependency{
				Module: &depModule,
				Output: marker.Output,
			})
			modTf, ok := depModule.Self().(*TFModule)
			if ok {
				modTf.ExpectedOutputs[marker.Output] = true
			} else {
				log.Debugf("Ignore adding outputs to module '%v', unknown type.", module.Name())
			}

			return reflect.ValueOf(resString), nil
		}
	}
	return subVal, nil
}
