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

const infraObjKindKey = "infrastructure"

type Infrastructure struct {
	ProjectPtr  *Project
	Backend     Backend
	Name        string
	BackendName string
	TemplateSrc string
	TemplateDir string
	Templates   []InfraTemplate
	Variables   map[string]interface{}
	ConfigData  map[string]interface{}
}

type infrastructureState struct {
}

func (p *Project) readInfrastructures() error {
	// Read and parse infrastructures.
	infras, exists := p.objects[infraObjKindKey]
	if !exists {
		err := fmt.Errorf("no infrastructures found, at least one needed")
		log.Debug(err.Error())
		return err
	}
	for _, infra := range infras {
		err := p.readInfrastructureObj(infra)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *Project) readInfrastructureObj(infraSpec ObjectData) error {
	name, ok := infraSpec.data["name"].(string)
	if !ok {
		return fmt.Errorf("infrastructure object must contain field 'name'")
	}
	// Check if infra with this name is already exists in project.
	if _, ok = p.Infrastructures[name]; ok {
		return fmt.Errorf("duplicate infrastructure name '%s'", name)
	}

	infra := Infrastructure{
		ProjectPtr: p,
		ConfigData: infraSpec.data,
		Name:       name,
	}

	infra.ConfigData["secret"], _ = p.configData["secret"]

	tmplSource, ok := infraSpec.data["template"].(string)
	if !ok {
		return fmt.Errorf("infrastructure object must contain field 'template'")
	}
	infra.Variables, ok = infraSpec.data["variables"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("infrastructure object must contain field 'variables'")
	}
	err := infra.ReadTemplates(tmplSource)
	if err != nil {
		return err
	}

	// Read backend name.
	infra.BackendName, ok = infraSpec.data["backend"].(string)
	if !ok {
		return fmt.Errorf("infrastructure object must contain field 'backend'")
	}
	bPtr, exists := p.Backends[infra.BackendName]
	if !exists {
		return fmt.Errorf("backend '%s' not found, infra: '%s'", infra.BackendName, infra.Name)
	}
	infra.Backend = bPtr
	p.Infrastructures[name] = &infra
	log.Debugf("Infrastructure added: %v", name)
	return nil
}

// ReadTemplates read all templates in src.
func (i *Infrastructure) ReadTemplates(src string) (err error) {
	// Read infra template data and apply variables.
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
		i.TemplateDir = templatesDir
	} else {
		templatesDownloadDir := filepath.Join(config.Global.TmpDir, "templates")
		os.Mkdir(templatesDownloadDir, os.ModePerm)
		dr, err := utils.GetTemplate(src, templatesDownloadDir, i.Name)
		if err != nil {
			return fmt.Errorf("download template: %v\n   See details about infrastructure template reference: https://github.com/shalb/cluster.dev#infra_options_template", err.Error())
		}
		i.TemplateDir, err = filepath.Rel(config.Global.WorkingDir, dr)
		if err != nil {
			return fmt.Errorf("reading templates: error parsing tmpl dir")
		}
	}

	templatesFilesList, err := filepath.Glob(i.TemplateDir + "/*.yaml")
	if err != nil {
		return err
	}
	i.Templates = []InfraTemplate{}
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
		infraTemplate, err := NewInfraTemplate(template)
		if err != nil {
			log.Debugf("reading templates: %v", err.Error())
			return err
		}
		i.Templates = append(i.Templates, *infraTemplate)
	}
	if len(i.Templates) < 1 {
		return fmt.Errorf("reading templates: no templates found")
	}
	i.TemplateSrc = src
	return nil
}

// TemplateTry apply infrastructure variables to template data.
// If template has unresolved variables - function will return an error.
func (i *Infrastructure) TemplateMust(data []byte) (res []byte, err error) {
	return i.tmplWithMissingKey(data, "error")
}

// TemplateTry apply infrastructure variables to template data.
// If template has unresolved variables - warn will be set to true.
func (i *Infrastructure) TemplateTry(data []byte) (res []byte, warn bool, err error) {
	res, err = i.tmplWithMissingKey(data, "default")
	if err != nil {
		return res, false, err
	}
	_, missingKeysErr := i.tmplWithMissingKey(data, "error")
	return res, missingKeysErr != nil, missingKeysErr
}

func (i *Infrastructure) tmplWithMissingKey(data []byte, missingKey string) (res []byte, err error) {
	tmpl, err := template.New("main").Funcs(i.ProjectPtr.TmplFunctionsMap).Option("missingkey=" + missingKey).Parse(string(data))
	if err != nil {
		return
	}
	templatedConf := bytes.Buffer{}
	err = tmpl.Execute(&templatedConf, i.ConfigData)
	return templatedConf.Bytes(), err
}
