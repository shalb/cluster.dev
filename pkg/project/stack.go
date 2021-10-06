package project

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"

	"github.com/apex/log"
	"github.com/shalb/cluster.dev/pkg/config"
	"github.com/shalb/cluster.dev/pkg/utils"
)

const stackObjKindKey = "Stack"

// Stack represent stack object.
type Stack struct {
	ProjectPtr       *Project
	Backend          Backend
	Name             string
	BackendName      string
	TemplateSrc      string
	TemplateDir      string
	Templates        []stackTemplate
	Variables        map[string]interface{}
	ConfigData       map[string]interface{}
	TmplFunctionsMap template.FuncMap
}

type stackState struct {
}

func (p *Project) readStacks() error {
	// Read and parse stacks.
	stacks, exists := p.objects[stackObjKindKey]
	if !exists {
		stacks, exists = p.objects["Infrastructure"]
		if !exists {
			err := fmt.Errorf("no stacks found, at least one needed")
			log.Debug(err.Error())
			return err
		}
		log.Warnf("'Infrastructure' key is deprecated and will be remover in future releases. Use 'Stack' instead")
	}
	for _, stack := range stacks {
		err := p.readStackObj(stack)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *Project) readStackObj(stackSpec ObjectData) error {
	name, ok := stackSpec.data["name"].(string)
	if !ok {
		return fmt.Errorf("stack object must contain field 'name'")
	}
	// Check if stack with this name is already exists in project.
	if _, ok = p.Stack[name]; ok {
		return fmt.Errorf("duplicate stack name '%s'", name)
	}

	stack := Stack{
		ProjectPtr:       p,
		ConfigData:       stackSpec.data,
		Name:             name,
		TmplFunctionsMap: make(template.FuncMap),
	}

	// Copy project template functions and add stack based (like readFile and templateFile)
	for fName, f := range p.TmplFunctionsMap {
		stack.TmplFunctionsMap[fName] = f
	}
	fReader := tmplFileReader{
		stackPtr: &stack,
	}
	stack.TmplFunctionsMap["readFile"] = fReader.ReadFile
	stack.TmplFunctionsMap["templateFile"] = fReader.TemplateFile

	// Copy secrets from project for templating.
	stack.ConfigData["secret"], _ = p.configData["secret"]

	tmplSource, ok := stackSpec.data["template"].(string)
	if !ok {
		return fmt.Errorf("stack object must contain field 'template'")
	}
	stack.Variables, ok = stackSpec.data["variables"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("stack object must contain field 'variables'")
	}
	err := stack.ReadTemplates(tmplSource)
	if err != nil {
		return err
	}

	// Read backend name.
	stack.BackendName, ok = stackSpec.data["backend"].(string)
	if !ok {
		return fmt.Errorf("stack object must contain field 'backend'")
	}
	bPtr, exists := p.Backends[stack.BackendName]
	if !exists {
		return fmt.Errorf("backend '%s' not found, stack: '%s'", stack.BackendName, stack.Name)
	}
	stack.Backend = bPtr
	p.Stack[name] = &stack
	log.Debugf("Stack added: %v", name)
	return nil
}

// ReadTemplates read all templates in src.
func (i *Stack) ReadTemplates(src string) (err error) {
	// Read stack template data and apply variables.
	var templatesDir string
	if utils.IsLocalPath(src) {
		if utils.IsAbsolutePath(src) {
			templatesDir = src
		} else {
			templatesDir = filepath.Join(config.Global.WorkingDir, src)
		}
		isDir, err := utils.CheckDir(templatesDir)
		if err != nil {
			return err
		}
		if !isDir {
			return fmt.Errorf("reading templates: local source should be a dir")
		}
		log.Debugf("Template dir: %v", templatesDir)
		i.TemplateDir, err = filepath.Rel(config.Global.WorkingDir, templatesDir)
		if err != nil {
			i.TemplateDir = templatesDir
		}
	} else {
		os.Mkdir(config.Global.TemplatesCacheDir, os.ModePerm)
		dr, err := utils.GetTemplate(src, config.Global.TemplatesCacheDir, i.Name)
		if err != nil {
			return fmt.Errorf("download template: %v\n   See details about stack template reference: https://github.com/shalb/cluster.dev#stack_options_template", err.Error())
		}
		log.Debugf("Template dir: %v", dr)
		i.TemplateDir, err = filepath.Rel(config.Global.WorkingDir, dr)
		if err != nil {
			return fmt.Errorf("reading templates: error parsing tmpl dir")
		}
	}

	templatesFilesList, err := filepath.Glob(i.TemplateDir + "/*.yaml")
	if err != nil {
		return err
	}
	i.Templates = []stackTemplate{}
	for _, fn := range templatesFilesList {
		tmplData, err := ioutil.ReadFile(fn)
		if err != nil {
			return err
		}
		var errIsWarn bool
		template, errIsWarn, err := i.TemplateTry(tmplData)
		if err != nil {
			if !errIsWarn {
				log.Fatal(err.Error())
			}
		}
		stackTemplate, err := NewStackTemplate(template)
		if err != nil {
			log.Debugf("reading templates: %v", err.Error())
			return err
		}
		i.Templates = append(i.Templates, *stackTemplate)
	}
	if len(i.Templates) < 1 {
		return fmt.Errorf("reading templates: no templates found")
	}
	i.TemplateSrc = src
	return nil
}

// TemplateMust apply stack variables to template data.
// If template has unresolved variables - function will return an error.
func (i *Stack) TemplateMust(data []byte) (res []byte, err error) {
	return i.tmplWithMissingKey(data, "error")
}

// TemplateTry apply stack variables to template data.
// If template has unresolved variables - warn will be set to true.
func (i *Stack) TemplateTry(data []byte) (res []byte, warn bool, err error) {
	res, err = i.tmplWithMissingKey(data, "default")
	if err != nil {
		return res, false, err
	}
	_, missingKeysErr := i.tmplWithMissingKey(data, "error")
	return res, missingKeysErr != nil, missingKeysErr
}

func (i *Stack) tmplWithMissingKey(data []byte, missingKey string) (res []byte, err error) {

	tmpl, err := template.New("main").Funcs(i.ProjectPtr.TmplFunctionsMap).Option("missingkey=" + missingKey).Parse(string(data))
	if err != nil {
		return
	}
	templatedConf := bytes.Buffer{}
	err = tmpl.Execute(&templatedConf, i.ConfigData)
	return templatedConf.Bytes(), err
}
