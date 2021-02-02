package project

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig"
	"github.com/apex/log"
	"github.com/shalb/cluster.dev/pkg/config"
)

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

func getEnv(varName string) (string, error) {
	if envVal, ok := os.LookupEnv(varName); ok {
		return envVal, nil
	}
	return "", fmt.Errorf("'%v' does not exists", varName)
}

func workDir() string {
	return config.Global.WorkingDir
}

var templateFunctionsMap = template.FuncMap{
	"ReconcilerVersionTag": printVersion,
	"env":                  getEnv,
	"workDir":              workDir,
}

func init() {
	for key, val := range sprig.FuncMap() {
		templateFunctionsMap[key] = val
	}
}

// RegisterTemplateDriver register module template driver.
func RegisterTemplateDriver(drv TemplateDriver) {
	TemplateDriversMap[drv.Name()] = drv
}

type TemplateDriver interface {
	AddTemplateFunctions(*Project)
	Name() string
}

var TemplateDriversMap map[string]TemplateDriver = map[string]TemplateDriver{}
