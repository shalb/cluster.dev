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

// addRemoteStateMarker function for template. Add hash marker, witch will be replaced with desired remote state.
func (m *terraformTemplateFunctions) addRemoteStateMarker(path string) (string, error) {

	_, ok := m.projectPtr.Markers[RemoteStateMarkerCatName]
	if !ok {
		m.projectPtr.Markers[RemoteStateMarkerCatName] = map[string]*project.DependencyOutput{}
	}
	splittedPath := strings.Split(path, ".")
	if len(splittedPath) != 3 {
		return "", fmt.Errorf("bad dependency path")
	}
	dep := project.DependencyOutput{
		Module:     nil,
		StackName:  splittedPath[0],
		ModuleName: splittedPath[1],
		Output:     splittedPath[2],
	}
	marker := project.CreateMarker("remoteState", fmt.Sprintf("%s.%s.%s", splittedPath[0], splittedPath[1], splittedPath[2]))
	m.projectPtr.Markers[RemoteStateMarkerCatName].(map[string]*project.DependencyOutput)[marker] = &dep
	return fmt.Sprintf("%s", marker), nil
}

// GetTemplateDriver return template driver to add template functions into the project.
func (m *Unit) GetTemplateDriver() (string, project.TemplateDriver) {
	return "terraform", &TerraformTemplateDriver{}
}

func init() {
	drv := TerraformTemplateDriver{}
	project.RegisterTemplateDriver(&drv)
}
