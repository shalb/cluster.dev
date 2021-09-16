package common

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/apex/log"
	"github.com/shalb/cluster.dev/pkg/project"
	"github.com/shalb/cluster.dev/pkg/utils"
)

// RemoteStatesScanner - project scanner function, witch process dependencies markers in module data setted by AddRemoteStateMarker template function.
func (m *Module) RemoteStatesScanner(data reflect.Value, module project.Module) (reflect.Value, error) {
	var subVal = data
	if data.Kind() != reflect.String {
		subVal = reflect.ValueOf(data.Interface())

	}

	resString := subVal.String()
	depMarkers, ok := module.ProjectPtr().Markers[RemoteStateMarkerCatName]
	if !ok {
		return subVal, nil
	}
	//markersList := map[string]*project.Dependency{}
	markersList, ok := depMarkers.(map[string]*project.DependencyOutput)
	if !ok {
		err := utils.JSONInterfaceToType(depMarkers, &markersList)
		if err != nil {
			return reflect.ValueOf(nil), fmt.Errorf("remote state scanner: read dependency: bad type")
		}
	}

	for key, marker := range markersList {
		if strings.Contains(resString, key) {
			var InfraName string
			if marker.InfraName == "this" {
				InfraName = module.InfraName()
			} else {
				InfraName = marker.InfraName
			}

			modKey := fmt.Sprintf("%s.%s", InfraName, marker.ModuleName)
			// log.Warnf("Mod Key: %v", modKey)
			depModule, exists := module.ProjectPtr().Modules[modKey]
			if !exists {
				log.Fatalf("Depend module does not exists. Src: '%s.%s', depend: '%s'", module.InfraName(), module.Name(), modKey)
			}
			markerTmp := project.DependencyOutput{Module: depModule, ModuleName: marker.ModuleName, InfraName: InfraName, Output: marker.Output}
			*module.Dependencies() = append(*module.Dependencies(), &markerTmp)
			m.markers[key] = &markerTmp
			depModule.ExpectedOutputs()[marker.Output] = &project.DependencyOutput{
				Output: marker.Output,
			}
		}
	}
	return reflect.ValueOf(resString), nil
}
