package project

import (
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/apex/log"
)

const OutputMarkerCatName = "outputMarkers"

// CheckContainsMarkers - check if string contains any template markers.
func (p *Project) CheckContainsMarkers(data string, kinds ...string) bool {
	for markersKind, markersSet := range p.Markers {
		vl := reflect.ValueOf(markersSet)
		if vl.Kind() != reflect.Map {
			log.Fatal("Internal error.")
		}
		var checkNeed bool = false
		if len(kinds) > 0 {
			for _, k := range kinds {
				if markersKind == k {
					checkNeed = true
					break
				}
			}
		} else {
			checkNeed = true
		}
		if !checkNeed {
			break
		}
		for _, marker := range vl.MapKeys() {
			if strings.Contains(data, marker.String()) {
				return true
			}
		}
	}
	return false
}

func (p *Project) addOutputMarker(output string) (string, error) {
	_, ok := p.Markers[OutputMarkerCatName]
	if !ok {
		p.Markers[OutputMarkerCatName] = map[string]*Dependency{}
	}
	splittedPath := strings.Split(output, ".")
	if len(splittedPath) != 3 {
		return "", fmt.Errorf("bad dependency path")
	}
	dep := Dependency{
		Module:     nil,
		InfraName:  splittedPath[0],
		ModuleName: splittedPath[1],
		Output:     splittedPath[2],
	}
	marker := p.CreateMarker(OutputMarkerCatName)
	p.Markers[OutputMarkerCatName].(map[string]*Dependency)[marker] = &dep

	return fmt.Sprintf("%s", marker), nil
}

func getEnv(varName string) (string, error) {
	if envVal, ok := os.LookupEnv(varName); ok {
		return envVal, nil
	}
	return "", fmt.Errorf("'%v' does not exists", varName)
}
