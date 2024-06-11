package project

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/apex/log"
	"github.com/shalb/cluster.dev/internal/config"
	"github.com/shalb/cluster.dev/pkg/utils"
)

const stackObjKindKey = "Stack"

// Stack represent stack object.
type Stack struct {
	ProjectPtr  *Project
	Backend     Backend
	Name        string
	BackendName string
	TemplateSrc string
	TemplateDir string
	Templates   []stackTemplate
	Variables   map[string]interface{}
	ConfigData  map[string]interface{}
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
		log.Warnf("'Infrastructure' key is deprecated and will be removed in future releases. Use 'Stack' instead")
	}
	for _, stack := range stacks {
		err := p.readStackObj(stack)
		if err != nil {
			return err
		}
	}
	if len(p.Stacks) == 0 {
		return fmt.Errorf("no stacks found, at least one needed")
	}
	return nil
}

func (p *Project) readStackObj(stackSpec ObjectData) error {
	name, ok := stackSpec.data["name"].(string)
	if !ok {
		return fmt.Errorf("stack object must contain field 'name'")
	}
	disabledInt := stackSpec.data["disabled"]
	if disabledInt != nil {
		disabled, ok := disabledInt.(bool)
		if !ok {
			return fmt.Errorf("stack option 'disabled' should be bool, not %T", disabledInt)
		}
		if disabled {
			log.Debugf("stack '%v' is disabled, ignore", name)
			return nil
		}
	}

	// Check if stack with this name is already exists in project.
	if _, ok = p.Stacks[name]; ok {
		return fmt.Errorf("duplicate stack name '%s'", name)
	}

	stack := Stack{
		ProjectPtr: p,
		ConfigData: stackSpec.data,
		Name:       name,
	}

	// Copy secrets from project for templating.
	stack.ConfigData["secret"] = p.configData["secret"]

	tmplSource, ok := stackSpec.data["template"].(string)
	if !ok {
		return fmt.Errorf("stack object must contain field 'template'")
	}
	stack.Variables, ok = stackSpec.data["variables"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("stack object must contain field 'variables'")
	}
	err := stack.ReadTemplate(tmplSource)
	if err != nil {
		return err
	}

	// Read backend name.
	stack.BackendName, ok = stackSpec.data["backend"].(string)
	if !ok {
		return fmt.Errorf("backend is not defined. To use default local backend, set 'backend: default' option")
	}
	bPtr, exists := p.Backends[stack.BackendName]
	if !exists {
		return fmt.Errorf("backend '%s' not found, stack: '%s'", stack.BackendName, stack.Name)
	}
	stack.Backend = bPtr
	p.Stacks[name] = &stack
	log.Debugf("Stack added: %v", name)
	return nil
}

// ReadTemplate read all templates in src.
func (s *Stack) ReadTemplate(src string) (err error) {
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
		s.TemplateDir, err = filepath.Rel(config.Global.WorkingDir, templatesDir)
		if err != nil {
			s.TemplateDir = templatesDir
		}
	} else {
		os.Mkdir(config.Global.TemplatesCacheDir, os.ModePerm)
		parsedRepoURL, err := utils.ParseGitUrl(src)
		if err != nil {
			return fmt.Errorf("download template: %w", err)
		}
		folderName, err := utils.URLToFolderName(parsedRepoURL.RepoString)
		if err != nil {
			return fmt.Errorf("download template: %w", err)
		}
		dr, err := utils.GetTemplate(src, config.Global.TemplatesCacheDir, folderName)
		if err != nil {
			return fmt.Errorf("download template: %w\n   See details about stack template reference: https://docs.cluster.dev/structure-stack/", err)
		}
		log.Debugf("Template dir: %v", dr)
		s.TemplateDir, err = filepath.Rel(config.Global.WorkingDir, dr)
		if err != nil {
			return fmt.Errorf("reading templates: error parsing tmpl dir: %w", err)
		}
	}

	templatesFilesList, err := filepath.Glob(s.TemplateDir + "/*.yaml")
	if err != nil {
		return err
	}
	s.Templates = []stackTemplate{}
	for _, fn := range templatesFilesList {
		tmplData, err := os.ReadFile(fn)
		if err != nil {
			return err
		}
		var errIsWarn bool
		template, errIsWarn, err := s.TemplateTry(tmplData, fn)
		if err != nil {
			if !errIsWarn {
				return err
			}
		}
		stackTemplate, err := NewStackTemplate(template)
		if err != nil {
			log.Debugf("reading templates: %v", err.Error())
			return err
		}
		s.Templates = append(s.Templates, *stackTemplate)
	}
	if len(s.Templates) < 1 {
		return fmt.Errorf("reading templates: no templates found")
	}
	s.TemplateSrc = src
	return nil
}

// TemplateMust apply stack variables to template data.
// If template has unresolved variables - function will return an error.
func (i *Stack) TemplateMust(data []byte, fileName string) (res []byte, err error) {
	return templateMust(data, i.ConfigData, i.ProjectPtr, i, fileName)
}

// TemplateTry apply values to template data.
// If template has unresolved variables - warn will be set to true.
func (i *Stack) TemplateTry(data []byte, fileName string) (res []byte, warn bool, err error) {
	return templateTry(data, i.ConfigData, i.ProjectPtr, i, fileName)
}
