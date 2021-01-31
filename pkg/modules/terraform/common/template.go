package common

import (
	"fmt"
	"strings"

	"github.com/apex/log"
	"github.com/shalb/cluster.dev/pkg/project"
)

type terraformTemplateFunctions struct {
	projectPtr *project.Project
}

type TerraformTemplateDriver struct {
}

func (m *TerraformTemplateDriver) AddTemplateFunctions(p *project.Project) {
	f := terraformTemplateFunctions{projectPtr: p}
	funcs := map[string]interface{}{
		"remoteState": f.addRemoteStateMarker,
		"insertYAML":  f.addYAMLBlockMarker,
	}
	for k, f := range funcs {
		_, ok := p.TmplFunctionsMap[k]
		if !ok {
			log.Debugf("Template Function '%v' added (terraform)", k)
			p.TmplFunctionsMap[k] = f
		}
	}
}

func (m *TerraformTemplateDriver) Name() string {
	return "terraform"
}

// RemoteStateMarkerCatName - name of markers category for remote states
const RemoteStateMarkerCatName = "RemoteStateMarkers"
const InsertYAMLMarkerCatName = "insertYAMLMarkers"

// addRemoteStateMarker function for template. Add hash marker, witch will be replaced with desired remote state.
func (m *terraformTemplateFunctions) addRemoteStateMarker(path string) (string, error) {

	_, ok := m.projectPtr.Markers[RemoteStateMarkerCatName]
	if !ok {
		m.projectPtr.Markers[RemoteStateMarkerCatName] = map[string]*project.Dependency{}
	}
	splittedPath := strings.Split(path, ".")
	if len(splittedPath) != 3 {
		return "", fmt.Errorf("bad dependency path")
	}
	dep := project.Dependency{
		Module:     nil,
		InfraName:  splittedPath[0],
		ModuleName: splittedPath[1],
		Output:     splittedPath[2],
	}
	marker := m.projectPtr.CreateMarker("remoteState")
	m.projectPtr.Markers[RemoteStateMarkerCatName].(map[string]*project.Dependency)[marker] = &dep
	return fmt.Sprintf("%s", marker), nil
}

// addYAMLBlockMarker function for template. Add hash marker, witch will be replaced with desired block.
func (m *terraformTemplateFunctions) addYAMLBlockMarker(data interface{}) (string, error) {
	_, ok := m.projectPtr.Markers[InsertYAMLMarkerCatName]
	if !ok {
		m.projectPtr.Markers[InsertYAMLMarkerCatName] = map[string]interface{}{}
	}
	marker := m.projectPtr.CreateMarker("YAML")
	m.projectPtr.Markers[InsertYAMLMarkerCatName].(map[string]interface{})[marker] = data
	return fmt.Sprintf("%s", marker), nil
}

// GetTemplateDriver return template driver to add template functions into the project.
func (m *Module) GetTemplateDriver() (string, project.TemplateDriver) {
	return "terraform", &TerraformTemplateDriver{}
}

func init() {
	drv := TerraformTemplateDriver{}
	project.RegisterTemplateDriver(&drv)
}
