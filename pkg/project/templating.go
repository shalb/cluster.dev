package project

import (
	"bytes"
	"fmt"
	"os"
	"reflect"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig"
	"github.com/apex/log"
	"github.com/shalb/cluster.dev/pkg/config"
	"golang.org/x/crypto/bcrypt"
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
	"reqEnv":               getEnv,
	"workDir":              workDir,
	"bcrypt":               BcryptString,
}

func init() {
	for key, val := range sprig.FuncMap() {
		if _, ok := templateFunctionsMap[key]; !ok {
			templateFunctionsMap[key] = val
		} else {
			log.Fatalf("Template functions name conflict '%v'", key)
		}
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

// TemplateMust do template
func (p *Project) TemplateMust(data []byte) (res []byte, err error) {
	return p.tmplWithMissingKey(data, "error")
}

// TemplateTry do template
func (p *Project) TemplateTry(data []byte) (res []byte, warn bool, err error) {
	res, err = p.tmplWithMissingKey(data, "default")
	if err != nil {
		return res, false, err
	}
	_, missingKeysErr := p.tmplWithMissingKey(data, "error")
	return res, missingKeysErr != nil, missingKeysErr
}

func (p *Project) tmplWithMissingKey(data []byte, missingKey string) (res []byte, err error) {
	tmpl, err := template.New("main").Funcs(p.TmplFunctionsMap).Option("missingkey=" + missingKey).Parse(string(data))
	if err != nil {
		return
	}
	templatedConf := bytes.Buffer{}
	err = tmpl.Execute(&templatedConf, p.configData)
	return templatedConf.Bytes(), err
}

func BcryptString(pwd []byte) (string, error) {

	hash, err := bcrypt.GenerateFromPassword(pwd, bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}
