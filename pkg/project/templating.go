package project

import (
	"bytes"
	"fmt"
	"os"
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
	for marker := range p.UnitLinks.Map() {
		if strings.Contains(data, marker) {
			return true
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
	AddTemplateFunctions(template.FuncMap, *Project, *Stack)
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
	tmplFuncMap := template.FuncMap{}
	// Copy common template functions.
	for k, v := range templateFunctionsMap {
		tmplFuncMap[k] = v
	}
	for _, drv := range TemplateDriversMap {
		drv.AddTemplateFunctions(tmplFuncMap, p, nil)
	}
	tmpl, err := template.New("main").Funcs(tmplFuncMap).Option("missingkey=" + missingKey).Parse(string(data))
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

const OutputLinkType = "outputMarkers"

// InsertYaml function for template. Insert data to yaml in json one line string (supported from a box by yaml unmarshal functions).
func InsertYaml(data interface{}) (string, error) {
	return utils.JSONEncodeString(data)
}

type GlobalTemplateDriver struct {
}

func (d *GlobalTemplateDriver) AddTemplateFunctions(mp template.FuncMap, p *Project, s *Stack) {

	addOutputMarkerFunk := func(path string) (string, error) {
		splittedPath := strings.Split(path, ".")
		// if len(splittedPath) != 3 {
		// 	return "", fmt.Errorf("bad dependency path 2")
		// }

		dep := ULinkT{
			Unit:            nil,
			LinkType:        OutputLinkType,
			TargetStackName: splittedPath[0],
			TargetUnitName:  splittedPath[1],
			OutputName:      splittedPath[2],
		}
		if s == nil && dep.TargetStackName == "this" {
			return "", fmt.Errorf("output tmpl: using 'this' allowed only in template, use stack name instead")
		}
		if dep.TargetStackName == "this" {
			dep.TargetStackName = s.Name
		}
		return p.UnitLinks.Set(&dep)
	}

	funcs := map[string]interface{}{
		"insertYAML": InsertYaml,
		"output":     addOutputMarkerFunk,
	}
	for k, f := range funcs {
		_, ok := mp[k]
		if !ok {
			// log.Debugf("Template Function '%v' added (%v)", k, d.Name())
			mp[k] = f
		}
	}
}

func (d *GlobalTemplateDriver) Name() string {
	return "global"
}

func init() {
	drv := GlobalTemplateDriver{}
	RegisterTemplateDriver(&drv)
}
