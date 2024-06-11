package base

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/shalb/cluster.dev/internal/project"
)

// RemoteStatesScanner - project scanner function, witch process dependencies markers in unit data setted by AddRemoteStateMarker template function.
func (u *Unit) RemoteStatesScanner(data reflect.Value, unit project.Unit) (reflect.Value, error) {
	var subVal = data
	if data.Kind() != reflect.String {
		subVal = reflect.ValueOf(data.Interface())
	}

	resString := subVal.String()

	markersList := u.ProjectPtr.UnitLinks.ByLinkTypes(RemoteStateLinkType).Map()

	for marker, link := range markersList {
		if strings.Contains(resString, marker) {
			var stackName string
			if link.TargetStackName == "this" {
				stackName = unit.Stack().Name
			} else {
				stackName = link.TargetStackName
			}

			modKey := fmt.Sprintf("%s.%s", stackName, link.TargetUnitName)
			depUnit, exists := unit.Project().Units[modKey]
			if !exists {
				return reflect.ValueOf(nil), fmt.Errorf("Depend unit does not exists. Src: '%s.%s', depend: '%s'", unit.Stack().Name, unit.Name(), modKey)
			}
			if link.Unit == nil {
				link.Unit = depUnit
			}
			// Add unit dependency link.
			if unit.Dependencies().Get(marker) == nil {
				unit.Dependencies().Set(link)
			}
		}
	}
	return reflect.ValueOf(resString), nil
}

// StringRemStScanner scan state data for outputs markers and replaces them for placeholders with remote state ref like <remoteState "stack.unit.output" >
func StringRemStScanner(data reflect.Value, unit project.Unit) (reflect.Value, error) {
	var subVal = data
	if data.Kind() != reflect.String {
		subVal = reflect.ValueOf(data.Interface())
	}
	resString := subVal.String()

	markersList := unit.Project().UnitLinks.ByLinkTypes(RemoteStateLinkType).Map()

	for key, marker := range markersList {
		if strings.Contains(resString, key) {
			resString = strings.ReplaceAll(resString, key, fmt.Sprintf("<remoteState %v.%v.%v>", marker.TargetStackName, marker.TargetUnitName, marker.OutputName))
		}
	}
	return reflect.ValueOf(resString), nil
}
