package base

import (
	"fmt"
	"strings"
	"text/template"

	"github.com/shalb/cluster.dev/pkg/project"
)

type terraformTemplateFunctions struct {
	projectPtr *project.Project
}

type TerraformTemplateDriver struct {
}

func (d *TerraformTemplateDriver) AddTemplateFunctions(mp template.FuncMap, p *project.Project, s *project.Stack) {
	addRemoteStateMarker := func(path string) (string, error) {
		splittedPath := strings.Split(path, ".")
		// if len(splittedPath) != 3 {
		// 	//return "", fmt.Errorf("bad dependency path 1")

		// }
		dep := project.ULinkT{
			Unit:            nil,
			LinkType:        RemoteStateLinkType,
			TargenStackName: splittedPath[0],
			TargetUnitName:  splittedPath[1],
			OutputName:      splittedPath[2],
		}
		if len(splittedPath) != 3 {
			dep.OutputName = strings.Join(splittedPath[2:], ".")
		}
		if s == nil && dep.TargenStackName == "this" {
			return "", fmt.Errorf("remoteState tmpl: using 'this' allowed only in template, use stack name instead")
		}
		if dep.TargenStackName == "this" {
			dep.TargenStackName = s.Name
		}
		return p.UnitLinks.Set(&dep)

	}

	funcs := map[string]interface{}{
		"remoteState": addRemoteStateMarker,
	}
	for k, f := range funcs {
		_, ok := mp[k]
		if !ok {
			// log.Debugf("Template Function '%v' added (%v)", k, d.Name())
			mp[k] = f
		}
	}
}

func (m *TerraformTemplateDriver) Name() string {
	return "terraform-new"
}

// GetTemplateDriver return template driver to add template functions into the project.
func (u *Unit) GetTemplateDriver() (string, project.TemplateDriver) {
	return "terraform", &TerraformTemplateDriver{}
}

func init() {
	drv := TerraformTemplateDriver{}
	project.RegisterTemplateDriver(&drv)
}
