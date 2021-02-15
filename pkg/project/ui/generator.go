package ui

import (
	"bytes"
	"embed"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"text/template"

	"github.com/apex/log"
	"github.com/paulrademacher/climenu"
	"github.com/shalb/cluster.dev/pkg/config"
	"github.com/shalb/cluster.dev/pkg/project"
	"gopkg.in/yaml.v3"
)

//go:embed templates/*
var templates embed.FS

const templatesDir = "templates"

func Test(p *project.Project) {
	log.Infof("Project: %v", p.Name())
}

func CreateProject() {
	prjTmplList, err := getProjectsTemplates()
	if err != nil {
		log.Fatalf("creating project: %v", err.Error())
	}
	menu := climenu.NewButtonMenu("Create project.", "Choose an project template:")
	generatorSpecs := map[string]tmplConfSpec{}
	for _, tmplName := range prjTmplList {
		sp, err := getGeneratorSpec(tmplName)
		if err != nil {
			log.Fatal(err.Error())
		}
		menu.AddMenuItem(sp.Description, tmplName)
		generatorSpecs[tmplName] = sp
	}
	action, escaped := menu.Run()
	dataForTemplating := map[string]string{}
	for _, opt := range generatorSpecs[action].Options {
		if opt.Regex == "" {
			opt.Regex = ".*"
		}
		for {
			respond := climenu.GetText(opt.Desciption, "")
			if verifyRegex(respond, opt.Regex) {
				dataForTemplating[opt.Name] = respond
				break
			}
			log.Warnf("Data verifycation error, regex fo check: '%v'", opt.Regex)
		}
	}

	if escaped {
		log.Warn("Exiting... ")
		os.Exit(0)
	}
	err = compileTree(filepath.Join(templatesDir, action, "data"), dataForTemplating)
	if err != nil {
		log.Fatalf("Create project:", err.Error())
	}
}

func compileTree(path string, optData interface{}, relPath ...string) (err error) {
	dir, err := templates.ReadDir(path)
	if err != nil {
		return
	}
	for _, elem := range dir {
		if elem.IsDir() {
			if err != nil {
				return
			}
			err = compileTree(filepath.Join(path, elem.Name()), optData, append(relPath, elem.Name())...)
			continue
		}

		fileTmpl, err := templates.ReadFile(filepath.Join(path, elem.Name()))
		if err != nil {
			return fmt.Errorf("internal error, %v", err.Error())
		}
		err = renderFile(fileTmpl, filepath.Join(filepath.Join(relPath...), elem.Name()), optData)
		if err != nil {
			return err
		}
	}
	return
}

func renderFile(tmplRaw []byte, outputFileName string, data interface{}) (err error) {
	tmpl, err := template.New("main").Delims("/{", "}/").Option("missingkey=error").Parse(string(tmplRaw))
	if err != nil {
		return
	}
	templatedConf := bytes.Buffer{}
	err = tmpl.Execute(&templatedConf, data)
	if err != nil {
		return
	}
	filename := filepath.Join(config.Global.WorkingDir, outputFileName)
	fileDir := filepath.Join(config.Global.WorkingDir, filepath.Dir(outputFileName))
	err = os.MkdirAll(fileDir, os.ModePerm)
	if err != nil {
		return
	}
	ioutil.WriteFile(filename, templatedConf.Bytes(), fs.ModePerm)
	if err != nil {
		return
	}
	log.Infof("Creating: %v", filepath.Base(outputFileName))
	return
}

func getProjectsTemplates() (res []string, err error) {
	dir, err := templates.ReadDir(templatesDir)
	if err != nil {
		return
	}
	for _, elem := range dir {
		if !elem.IsDir() {
			err = fmt.Errorf("reading templates: internal error")
			return
		}
		res = append(res, elem.Name())
	}
	return
}

type optsSpec struct {
	Name       string `yaml:"name"`
	Desciption string `yaml:"description"`
	Regex      string `yaml:"regex,omitempty"`
}

type tmplConfSpec struct {
	Description string     `yaml:"description"`
	Options     []optsSpec `yaml:"options"`
}

func getGeneratorSpec(templateName string) (res tmplConfSpec, err error) {
	configRaw, err := templates.ReadFile(filepath.Join(templatesDir, templateName, "config.yaml"))
	if err != nil {
		err = fmt.Errorf("reading template: internal error (file not found %v): %v", filepath.Join(templatesDir, templateName, "config.yaml"), err.Error())
		return
	}
	rs := tmplConfSpec{}
	err = yaml.Unmarshal(configRaw, &rs)
	res = rs
	return
}

func verifyRegex(name, regex string) bool {
	return regexp.MustCompile(regex).MatchString(name)
}
