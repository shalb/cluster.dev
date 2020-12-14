package reconciler

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
)

type Infrastructure struct {
	Name      string
	Backend   string
	Template  []byte
	Variables map[string]interface{}
}

func (g *Project) readInfrastructureObj(obj map[string]interface{}) error {
	name, ok := obj["name"].(string)
	if !ok {
		return fmt.Errorf("infrastructure object must contain field 'name'")
	}
	// Check if infra with this name is already exists in project.
	if _, ok = g.Infrastructures[name]; ok {
		return fmt.Errorf("Duplicate infrastructure name '%s'", name)
	}

	infra := Infrastructure{}
	tmplFileName, ok := obj["template"].(string)
	if !ok {
		return fmt.Errorf("infrastructure object must contain field 'template'")
	}

	infra.Variables, ok = obj["variables"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("infrastructure object must contain field 'variables'")
	}

	// Read infra template data and apply variables.
	tmplData, err := ioutil.ReadFile(tmplFileName)
	if err != nil {
		return err
	}

	t, err := template.New("main").Funcs(g.TmplFunctionsMap).Option("missingkey=error").Parse(string(tmplData))

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
	infra.Backend, ok = obj["backend"].(string)
	if !ok {
		return fmt.Errorf("infrastructure object must contain field 'provider'")
	}
	infra.Name = name
	g.Infrastructures[name] = &infra
	return nil
}
