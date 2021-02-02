package project

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"text/template"

	"github.com/apex/log"
	"github.com/shalb/cluster.dev/pkg/config"
)

const infraObjKindKey = "infrastructure"

type Infrastructure struct {
	ProjectPtr  *Project
	Backend     Backend
	Name        string
	BackendName string
	TemplateSrc string
	Template    []byte
	Variables   map[string]interface{}
	FullSpec    map[string]interface{}
}

func (i *Infrastructure) DoTemplate(in []byte) ([]byte, error) {
	t, err := template.New("main").Funcs(i.ProjectPtr.TmplFunctionsMap).Option("missingkey=default").Parse(string(in))

	if err != nil {
		return nil, err
	}

	tmpl := bytes.Buffer{}
	err = t.Execute(&tmpl, i.FullSpec)
	if err != nil {
		return nil, err
	}
	return tmpl.Bytes(), nil
}

func (p *Project) readInfrastructureObj(infraSpec map[string]interface{}) error {
	name, ok := infraSpec["name"].(string)
	if !ok {
		return fmt.Errorf("infrastructure object must contain field 'name'")
	}
	// Check if infra with this name is already exists in project.
	if _, ok = p.Infrastructures[name]; ok {
		return fmt.Errorf("Duplicate infrastructure name '%s'", name)
	}

	infra := Infrastructure{
		ProjectPtr: p,
		FullSpec:   infraSpec,
	}
	tmplFileName, ok := infraSpec["template"].(string)
	if !ok {
		return fmt.Errorf("infrastructure object must contain field 'template'")
	}

	infra.Variables, ok = infraSpec["variables"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("infrastructure object must contain field 'variables'")
	}
	// Read infra template data and apply variables.
	templatesFile := filepath.Join(config.Global.WorkingDir, tmplFileName)
	tmplData, err := ioutil.ReadFile(templatesFile)
	if err != nil {
		return err
	}

	infra.Template, err = infra.DoTemplate(tmplData)
	if err != nil {
		return err
	}

	// Read backend name.
	infra.BackendName, ok = infraSpec["backend"].(string)
	if !ok {
		return fmt.Errorf("infrastructure object must contain field 'backend'")
	}
	bPtr, exists := p.Backends[infra.BackendName]
	if !exists {
		return fmt.Errorf("Backend '%s' not found, infra: '%s'", infra.BackendName, infra.Name)
	}
	infra.Backend = bPtr
	infra.Name = name
	infra.TemplateSrc = tmplFileName
	p.Infrastructures[name] = &infra
	log.Infof("Infrastructure '%v' added", name)
	return nil
}
