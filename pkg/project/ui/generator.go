package ui

import (
	"embed"
	"fmt"
	"path/filepath"

	"github.com/apex/log"
	"github.com/shalb/cluster.dev/pkg/project"
)

//go:embed templates/*
var templates embed.FS

const templatesDir = "templates"

func Test(p *project.Project) {
	log.Infof("Project: %v", p.Name())
}

func Create() {
	prjTmplList, err := getProjectsTemplates()
	if err != nil {
		log.Fatalf("creating project: %v", err.Error())
	}
	for _, tmpl := range prjTmplList {
		log.Info(tmpl)
	}
}

func compileTree(path string) (err error) {
	dir, err := templates.ReadDir(path)
	if err != nil {
		return
	}
	for _, elem := range dir {
		if elem.IsDir() {
			err = compileTree(filepath.Join(path, elem.Name()))
		}
		if err != nil {
			return
		}
		log.Info(filepath.Join(path, elem.Name()))
	}
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

type varsSpec struct {
	Name       string `yaml:"name"`
	Desciption string `yaml:"description"`
}
