package project

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig"
	"github.com/apex/log"
	"github.com/shalb/cluster.dev/pkg/config"
	"github.com/shalb/cluster.dev/pkg/utils"
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
	"cidrSubnet":           utils.CidrSubnet,
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

// RegisterTemplateDriver register unit template driver.
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

type tmplFileReader struct {
	stackPtr *Stack
}

func (t tmplFileReader) ReadFile(path string) (string, error) {
	vfPath := filepath.Join(t.stackPtr.TemplateDir, path)
	valuesFileContent, err := ioutil.ReadFile(vfPath)
	if err != nil {
		log.Debugf(err.Error())
		return "", err
	}
	return string(valuesFileContent), nil
}

func (t tmplFileReader) TemplateFile(path string) (string, error) {
	vfPath := filepath.Join(t.stackPtr.TemplateDir, path)
	rawFile, err := ioutil.ReadFile(vfPath)
	if err != nil {
		log.Debugf(err.Error())
		return "", err
	}
	templatedFile, errIsWarn, err := t.stackPtr.TemplateTry(rawFile)
	if err != nil {
		if !errIsWarn {
			log.Fatal(err.Error())
		}
	}
	return string(templatedFile), nil
}

const OutputMarkerCatName = "outputMarkers"

// insertYaml function for template. Add hash marker, witch will be replaced with desired block.
func (p *Project) insertYaml(data interface{}) (string, error) {
	return utils.JSONEncodeString(data)
}

type GlobalTemplateDriver struct {
}

func (d *GlobalTemplateDriver) AddTemplateFunctions(p *Project) {
	funcs := map[string]interface{}{
		"insertYAML": p.insertYaml,
		"output":     p.addOutputMarker,
	}
	for k, f := range funcs {
		_, ok := p.TmplFunctionsMap[k]
		if !ok {
			log.Debugf("Template Function '%v' added (%v)", k, d.Name())
			p.TmplFunctionsMap[k] = f
		}
	}
}

func (d *GlobalTemplateDriver) Name() string {
	return "global"
}

// addOutputMarker function for template. Add hash marker, witch will be replaced with desired unit output.
func (p *Project) addOutputMarker(path string) (string, error) {

	_, ok := p.Markers[OutputMarkerCatName]
	if !ok {
		p.Markers[OutputMarkerCatName] = map[string]*DependencyOutput{}
	}
	splittedPath := strings.Split(path, ".")
	if len(splittedPath) != 3 {
		return "", fmt.Errorf("bad dependency path")
	}
	dep := DependencyOutput{
		Unit:      nil,
		StackName: splittedPath[0],
		UnitName:  splittedPath[1],
		Output:    splittedPath[2],
	}
	marker := CreateMarker("output", fmt.Sprintf("%s.%s.%s", splittedPath[0], splittedPath[1], splittedPath[2]))
	p.Markers[OutputMarkerCatName].(map[string]*DependencyOutput)[marker] = &dep
	return fmt.Sprintf("%s", marker), nil
}

// OutputsScanner - project scanner function, witch process dependencies markers in unit data setted by AddRemoteStateMarker template function.
func OutputsScanner(data reflect.Value, unit Unit) (reflect.Value, error) {
	var subVal = data
	if data.Kind() != reflect.String {
		subVal = reflect.ValueOf(data.Interface())
	}
	resString := subVal.String()
	markersList := map[string]*DependencyOutput{}
	err := unit.Project().GetMarkers(OutputMarkerCatName, &markersList)
	if err != nil {
		return reflect.ValueOf(nil), fmt.Errorf("process outputs: %w", err)
	}
	for key, marker := range markersList {
		if strings.Contains(resString, key) {
			if marker.StackName == "this" {
				marker.StackName = unit.Stack().Name
			}
			modKey := fmt.Sprintf("%s.%s", marker.StackName, marker.UnitName)
			depUnit, exists := unit.Project().Units[modKey]
			if !exists {
				log.Fatalf("Depend unit does not exists. Src: '%s.%s', depend: '%s'", unit.Stack().Name, unit.Name(), modKey)
			}
			o, exists := depUnit.ExpectedOutputs()[marker.Output]
			if exists && o.OutputData != "" {
				resString = strings.ReplaceAll(resString, key, o.OutputData)
				return reflect.ValueOf(resString), nil
			}
			outputTmp := marker
			if unit.FindDependency(outputTmp.StackName, outputTmp.UnitName) == nil {
				*unit.Dependencies() = append(*unit.Dependencies(), outputTmp)
			}
			depUnit.ExpectedOutputs()[marker.Output] = outputTmp
		}
	}
	return reflect.ValueOf(resString), nil
}

// // OutputsScanner - project scanner function, witch process dependencies markers in unit data setted by AddRemoteStateMarker template function.
// func OutputsScannerDebug(data reflect.Value, unit Unit) (reflect.Value, error) {
// 	var subVal = data
// 	if data.Kind() != reflect.String {
// 		subVal = reflect.ValueOf(data.Interface())
// 	}
// 	resString := subVal.String()
// 	markersList := map[string]*DependencyOutput{}
// 	err := unit.Project().GetMarkers(OutputMarkerCatName, &markersList)
// 	if err != nil {
// 		return reflect.ValueOf(nil), fmt.Errorf("process outputs: %w", err)
// 	}
// 	for key, marker := range markersList {
// 		if strings.Contains(resString, key) {
// 			if marker.StackName == "this" {
// 				marker.StackName = unit.Stack().Name
// 			}
// 			modKey := fmt.Sprintf("%s.%s", marker.StackName, marker.UnitName)
// 			depUnit, exists := unit.Project().Units[modKey]
// 			if !exists {
// 				log.Fatalf("Depend unit does not exists. Src: '%s.%s', depend: '%s'", unit.Stack().Name, unit.Name(), modKey)
// 			}
// 			o, exists := depUnit.ExpectedOutputs()[marker.Output]
// 			log.Warnf("Output Marker found: %v, Output data: %+v", key, o)
// 			if exists && o.OutputData != "" {
// 				resString = strings.ReplaceAll(resString, key, o.OutputData)
// 				return reflect.ValueOf(resString), nil
// 			}
// 			outputTmp := marker
// 			if unit.FindDependency(outputTmp.StackName, outputTmp.UnitName) == nil {
// 				*unit.Dependencies() = append(*unit.Dependencies(), outputTmp)
// 			}
// 			depUnit.ExpectedOutputs()[marker.Output] = outputTmp
// 		}
// 	}
// 	return reflect.ValueOf(resString), nil
// }

// StateOutputsScanner scan state data for outputs markers and replaces them for placeholders with output ref like <output "stack.unit.output" >
func StateOutputsScanner(data reflect.Value, unit Unit) (reflect.Value, error) {
	var subVal = data
	if data.Kind() != reflect.String {
		subVal = reflect.ValueOf(data.Interface())
	}
	resString := subVal.String()
	depMarkers, ok := unit.Project().Markers[OutputMarkerCatName]
	if !ok {
		return subVal, nil
	}
	//markersList := map[string]*project.Dependency{}
	markersList, ok := depMarkers.(map[string]*DependencyOutput)
	if !ok {
		err := utils.JSONCopy(depMarkers, &markersList)
		if err != nil {
			return reflect.ValueOf(nil), fmt.Errorf("remote state scanner: read dependency: bad type")
		}
	}

	for key, marker := range markersList {
		if strings.Contains(resString, key) {
			resString = strings.ReplaceAll(resString, key, fmt.Sprintf("<output %v.%v.%v>", marker.StackName, marker.UnitName, marker.Output))
		}
	}
	return reflect.ValueOf(resString), nil
}

func init() {
	drv := GlobalTemplateDriver{}
	RegisterTemplateDriver(&drv)
}
