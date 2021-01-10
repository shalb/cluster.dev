package project

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"path/filepath"

	"github.com/apex/log"
	"github.com/shalb/cluster.dev/internal/config"
)

const infraObjKindKey = "infrastructure"

type Infrastructure struct {
	ProjectPtr  *Project
	Backend     Backend
	Name        string
	BackendName string
	Template    []byte
	Variables   map[string]interface{}
}

func (p *Project) readInfrastructureObj(obj map[string]interface{}) error {
	name, ok := obj["name"].(string)
	if !ok {
		return fmt.Errorf("infrastructure object must contain field 'name'")
	}
	// Check if infra with this name is already exists in project.
	if _, ok = p.Infrastructures[name]; ok {
		return fmt.Errorf("Duplicate infrastructure name '%s'", name)
	}

	infra := Infrastructure{
		ProjectPtr: p,
	}
	tmplFileName, ok := obj["template"].(string)
	if !ok {
		return fmt.Errorf("infrastructure object must contain field 'template'")
	}

	infra.Variables, ok = obj["variables"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("infrastructure object must contain field 'variables'")
	}

	// Read infra template data and apply variables.
	templatesFile := filepath.Join(config.Global.WorkingDir, tmplFileName)
	tmplData, err := ioutil.ReadFile(templatesFile)
	if err != nil {
		return err
	}

	t, err := template.New("main").Funcs(p.TmplFunctionsMap).Option("missingkey=error").Parse(string(tmplData))

	if err != nil {
		return err
	}

	tmpl := bytes.Buffer{}
	err = t.Execute(&tmpl, obj)
	if err != nil {
		return err
	}

	infra.Template = tmpl.Bytes()

	// Read backend name.
	infra.BackendName, ok = obj["backend"].(string)
	if !ok {
		return fmt.Errorf("infrastructure object must contain field 'backend'")
	}
	bPtr, exists := p.Backends[infra.BackendName]
	if !exists {
		return fmt.Errorf("Backend '%s' not found, infra: '%s'", infra.BackendName, infra.Name)
	}
	infra.Backend = bPtr
	infra.Name = name
	p.Infrastructures[name] = &infra
	log.Infof("Infrastructure '%v' added", name)
	return nil
}
