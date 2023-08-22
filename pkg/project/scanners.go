package project

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/apex/log"
)

// OutputsScanner - project scanner function, witch process dependencies markers in unit data setted by AddRemoteStateMarker template function.
func OutputsScanner(data reflect.Value, unit Unit) (reflect.Value, error) {
	var subVal = data
	if data.Kind() != reflect.String {
		subVal = reflect.ValueOf(data.Interface())
	}
	resString := subVal.String()
	markersList := unit.Project().UnitLinks.ByLinkTypes(OutputLinkType).Map()
	for marker, link := range markersList {
		if strings.Contains(resString, marker) {

			if link.TargetStackName == "this" {
				link.TargetStackName = unit.Stack().Name

			}
			modKey := fmt.Sprintf("%s.%s", link.TargetStackName, link.TargetUnitName)
			depUnit, exists := unit.Project().Units[modKey]
			if !exists {
				log.Fatalf("Depend unit does not exists. Src: '%s.%s', depend: '%s'", unit.Stack().Name, unit.Name(), modKey)
			}
			// Add unit ptr to unit link.
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

// OutputsReplacer - project scanner function, witch process dependencies markers in unit data setted by AddRemoteStateMarker template function.
func OutputsReplacer(data reflect.Value, unit Unit) (reflect.Value, error) {
	var subVal = data
	if data.Kind() != reflect.String {
		subVal = reflect.ValueOf(data.Interface())
	}
	resString := subVal.String()
	markersList := unit.Project().UnitLinks.ByLinkTypes(OutputLinkType).Map()
	for marker, link := range markersList {
		if strings.Contains(resString, marker) {

			if link.Unit == nil {
				return reflect.ValueOf(nil), fmt.Errorf("replace output internal error: unit link does not initted")
			}
			if link.OutputData == nil {
				log.Warnf("The output data is unavalible. Inserting placeholder <output %s.%s>.", link.TargetStackName, link.TargetUnitName)
				resString = strings.ReplaceAll(resString, marker, fmt.Sprintf("<output %s.%s>", link.TargetStackName, link.TargetUnitName))
			}
			if resString == marker {
				return reflect.ValueOf(link.OutputData), nil
			} else {
				var dataStr string
				if reflect.ValueOf(link.OutputData).Kind() != reflect.String {
					// TODO process error if output data is not string.
					// For now - convert to string.
					dataStr = fmt.Sprintf("%v", link.OutputData)
				} else {
					dataStr = link.OutputData.(string)
				}
				resString = strings.ReplaceAll(resString, marker, dataStr)
			}
		}
	}
	return reflect.ValueOf(resString), nil
}

// StateOutputsReplacer scan state data for outputs markers and replaces them for placeholders with output ref like <output "stack.unit.output" >
func StateOutputsReplacer(data reflect.Value, unit Unit) (reflect.Value, error) {
	var subVal = data
	if data.Kind() != reflect.String {
		subVal = reflect.ValueOf(data.Interface())
	}
	resString := subVal.String()
	markersList := unit.Project().UnitLinks.ByLinkTypes(OutputLinkType).Map()
	for key, marker := range markersList {
		if strings.Contains(resString, key) {
			resString = strings.ReplaceAll(resString, key, fmt.Sprintf("<output %v.%v.%v>", marker.TargetStackName, marker.TargetUnitName, marker.OutputName))
		}
	}
	return reflect.ValueOf(resString), nil
}
