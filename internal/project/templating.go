package project

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig"
	"github.com/apex/log"
	"github.com/shalb/cluster.dev/internal/config"
	"github.com/shalb/cluster.dev/pkg/utils"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/yaml.v3"
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

func toYaml(in interface{}) (string, error) {
	res, err := yaml.Marshal(in)
	return string(res), err
}

// Type for tmpl functions list with additional options, like set path to templated file.
type ExtendedFuncMap template.FuncMap

func (f *ExtendedFuncMap) Get(path string, s *Stack) template.FuncMap {
	readFileEx := func(f string) (string, error) {
		return readFile(f, path)
	}
	var p, pathFuncName string
	if s == nil {
		pathFuncName = "projectPath"
		p = config.Global.ProjectConfigsPath
	} else {
		pathFuncName = "templatePath"
		p = filepath.Join(config.Global.ProjectConfigsPath, s.TemplateDir)
	}
	getPath := func() string {
		return p
	}
	var templateFunctionsMap = template.FuncMap{
		"ReconcilerVersionTag": printVersion,
		"reqEnv":               getEnv,
		"workDir":              workDir,
		"bcrypt":               BcryptString,
		"cidrSubnet":           utils.CidrSubnet,
		"toYaml":               toYaml,
		"readFile":             readFileEx,
		pathFuncName:           getPath,
	}
	for key, val := range sprig.FuncMap() {
		if _, ok := templateFunctionsMap[key]; !ok {
			templateFunctionsMap[key] = val
		} else {
			log.Fatalf("Template functions name conflict '%v'", key)
		}
	}
	return templateFunctionsMap
}

// readFile template function to read files in project folder.
func readFile(relativePath string, dir string) (string, error) {
	fullPath := filepath.Join(dir, relativePath)
	res, err := os.ReadFile(fullPath)
	return string(res), err
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
func (p *Project) TemplateMust(data []byte, fileName string) (res []byte, err error) {
	return templateMust(data, p.configData, p, nil, fileName)
}

// TemplateTry do template
func (p *Project) TemplateTry(data []byte, fileName string) (res []byte, warn bool, err error) {
	return templateTry(data, p.configData, p, nil, fileName)
}

// templateMust apply values to template data, considering template file path (if empty will be used project path).
// If template has unresolved variables - function will return an error.
func templateMust(data []byte, values interface{}, p *Project, s *Stack, fileName string) (res []byte, err error) {
	return tmplWithMissingKey(data, values, "error", p, s, fileName)
}

// templateMust apply values to template data, considering template file path (if empty will be used project path).
// If template has unresolved variables - warn will be set to true.
func templateTry(data []byte, values interface{}, p *Project, s *Stack, fileName string) (res []byte, warn bool, err error) {
	res, err = tmplWithMissingKey(data, values, "default", p, s, fileName)
	if err != nil {
		return res, false, err
	}
	_, missingKeysErr := tmplWithMissingKey(data, values, "error", p, s, fileName)
	return res, missingKeysErr != nil, missingKeysErr
}

// templateMust apply values to template data, considering template file path (if empty will be used project path).
// If use stack pointer for units functions integration.
func tmplWithMissingKey(data []byte, values interface{}, missingKey string, p *Project, s *Stack, fileName string) (res []byte, err error) {
	tmplFuncMap := template.FuncMap{}

	// If file path is relative - convert to absolute using project dir as base.
	if !utils.IsAbsolutePath(fileName) {
		fileName = filepath.Join(config.Global.ProjectConfigsPath, fileName)
	}
	// Copy common template functions.
	funcs := ExtendedFuncMap{}
	for k, v := range funcs.Get(filepath.Dir(fileName), s) {
		tmplFuncMap[k] = v
	}
	for _, drv := range TemplateDriversMap {
		drv.AddTemplateFunctions(tmplFuncMap, p, s)
	}
	tmpl, err := template.New("main").Funcs(tmplFuncMap).Option("missingkey=" + missingKey).Parse(string(data))
	if err != nil {
		return
	}
	templatedConf := bytes.Buffer{}
	err = tmpl.Execute(&templatedConf, values)
	// log.Warnf("tmplWithMissingKey file: %v", fileName)
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
		if len(splittedPath) == 2 {
			splittedPath = append([]string{"this"}, splittedPath...)
		}

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
