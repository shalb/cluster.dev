package common

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/apex/log"
	"github.com/shalb/cluster.dev/pkg/project"
)

// RemoteStatesScanner - project scanner function, witch process dependencies markers in module data setted by AddRemoteStateMarker template function.
func (m *Module) RemoteStatesScanner(data reflect.Value, module project.Module) (reflect.Value, error) {
	subVal := reflect.ValueOf(data.Interface())
	resString := subVal.String()
	depMarkers, ok := module.ProjectPtr().Markers[RemoteStateMarkerCatName]
	if !ok {
		return subVal, nil
	}
	for key, marker := range depMarkers.(map[string]*project.Dependency) {
		if strings.Contains(resString, key) {
			log.Debugf("SCANNER marker: %v", key)
			if marker.InfraName == "this" {
				marker.InfraName = module.InfraName()
			}
			modKey := fmt.Sprintf("%s.%s", marker.InfraName, marker.ModuleName)
			depModule, exists := module.ProjectPtr().Modules[modKey]
			if !exists {
				log.Fatalf("Depend module does not exists. Src: '%s.%s', depend: '%s'", module.InfraName(), module.Name(), modKey)
			}
			*module.Dependencies() = append(*module.Dependencies(), marker)
			depModule.ExpectedOutputs()[marker.Output] = true
			return reflect.ValueOf(resString), nil
		}
	}
	return subVal, nil
}
func (m *Module) YamlBlockMarkerScanner(data reflect.Value, module project.Module) (reflect.Value, error) {
	subVal := reflect.ValueOf(data.Interface())

	yamlMarkers, ok := module.ProjectPtr().Markers[InsertYAMLMarkerCatName].(map[string]interface{})
	if !ok {
		return subVal, nil
	}
	for hash := range yamlMarkers {
		if subVal.String() == hash {
			return reflect.ValueOf(yamlMarkers[hash]), nil
		}
	}
	return subVal, nil
}
