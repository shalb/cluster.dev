package base

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/apex/log"
	"github.com/shalb/cluster.dev/pkg/project"
	"github.com/shalb/cluster.dev/pkg/utils"
)

// RemoteStatesScanner - project scanner function, witch process dependencies markers in unit data setted by AddRemoteStateMarker template function.
func (m *Unit) RemoteStatesScanner(data reflect.Value, unit project.Unit) (reflect.Value, error) {
	var subVal = data
	if data.Kind() != reflect.String {
		subVal = reflect.ValueOf(data.Interface())

	}

	resString := subVal.String()
	depMarkers, ok := unit.Project().Markers[RemoteStateMarkerCatName]
	if !ok {
		return subVal, nil
	}
	//markersList := map[string]*project.Dependency{}
	markersList, ok := depMarkers.(map[string]*project.DependencyOutput)
	if !ok {
		err := utils.JSONCopy(depMarkers, &markersList)
		if err != nil {
			return reflect.ValueOf(nil), fmt.Errorf("remote state scanner: read dependency: bad type")
		}
	}

	for key, marker := range markersList {
		if strings.Contains(resString, key) {
			var stackName string
			if marker.StackName == "this" {
				stackName = unit.Stack().Name
			} else {
				stackName = marker.StackName
			}

			modKey := fmt.Sprintf("%s.%s", stackName, marker.UnitName)
			// log.Warnf("Mod Key: %v", modKey)
			depUnit, exists := unit.Project().Units[modKey]
			if !exists {
				log.Fatalf("Depend unit does not exists. Src: '%s.%s', depend: '%s'", unit.Stack().Name, unit.Name(), modKey)
			}
			markerTmp := project.DependencyOutput{Unit: depUnit, UnitName: marker.UnitName, StackName: stackName, Output: marker.Output}
			*unit.Dependencies() = append(*unit.Dependencies(), &markerTmp)
			m.UnitMarkers[key] = &markerTmp
			depUnit.ExpectedOutputs()[marker.Output] = &project.DependencyOutput{
				Output: marker.Output,
			}
		}
	}
	// log.Infof("%v", reflect.ValueOf(resString).Kind())
	return reflect.ValueOf(resString), nil
}
