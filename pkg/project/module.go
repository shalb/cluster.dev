package project

import (
	"bytes"
	"fmt"
	"html/template"
	"reflect"
	"regexp"
	"strings"

	"github.com/apex/log"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/shalb/cluster.dev/pkg/hcltools"
	"github.com/zclconf/go-cty/cty"
)

// Module - describe module.
type Module struct {
	InfraPtr     *infrastructure
	ProjectPtr   *Project
	BackendPtr   Backend
	Name         string
	Type         string
	Source       string
	Inputs       map[string]interface{}
	Dependencies []*Dependency
	Outputs      map[string]bool
}

// Dependency describe module dependency.
type Dependency struct {
	Module     *Module
	ModuleName string
	InfraName  string
	Output     string
}

// GenMainCodeBlockHCL generate main code block for this module.
func (m *Module) GenMainCodeBlockHCL() ([]byte, error) {
	f := hclwrite.NewEmptyFile()
	rootBody := f.Body()
	moduleBlock := rootBody.AppendNewBlock("module", []string{m.Name})
	moduleBody := moduleBlock.Body()
	moduleBody.SetAttributeValue("source", cty.StringVal(m.Source))
	for key, val := range m.Inputs {
		ctyVal, err := hcltools.InterfaceToCty(val)
		if err != nil {
			return nil, err
		}
		moduleBody.SetAttributeValue(key, ctyVal)
	}
	for hash, marker := range m.ProjectPtr.DependencyMarkers {
		remoteStateRef := fmt.Sprintf("data.terraform_remote_state.%s-%s.outputs.%s", marker.InfraName, marker.ModuleName, marker.Output)
		hcltools.ReplaceStingMarkerInBody(moduleBody, hash, remoteStateRef)
	}
	return f.Bytes(), nil
}

// GenBackendCodeBlock generate backend code block for this module.
func (m *Module) GenBackendCodeBlock() ([]byte, error) {

	res, err := m.BackendPtr.GetBackendHCL(*m)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// GenOutputsBlock generate output code block for this module.
func (m *Module) GenOutputsBlock() ([]byte, error) {
	f := hclwrite.NewEmptyFile()

	rootBody := f.Body()
	for output := range m.Outputs {
		log.Debugf("Output: %v", output)
		re := regexp.MustCompile(`^[A-Za-z][a-zA-Z0-9_\-]{0,}`)
		outputName := re.FindString(output)
		if len(outputName) < 1 {
			return nil, fmt.Errorf("invalid output '%v' in module '%v'", output, m.Name)
		}
		dataBlock := rootBody.AppendNewBlock("output", []string{outputName})
		dataBody := dataBlock.Body()
		outputStr := fmt.Sprintf("module.%s.%s", m.Name, outputName)
		dataBody.SetAttributeRaw("value", hcltools.CreateTokensForOutput(outputStr))
	}
	return f.Bytes(), nil

}

// GenDepsRemoteStates generate terraform remote states for all dependencies of this module.
func (m *Module) GenDepsRemoteStates() ([]byte, error) {
	var res []byte
	depsUniq := map[*Module]bool{}

	for _, dep := range m.Dependencies {
		if _, ok := depsUniq[dep.Module]; ok {
			continue
		}
		depsUniq[dep.Module] = true
		rs, err := dep.Module.GenRemoteStateToSelf()
		if err != nil {
			return nil, err
		}
		res = append(res, rs...)
	}
	return res, nil
}

// GenRemoteStateToSelf - remote state block generate terraform code. It's remote state !to this module! witch should be used in another module depend of this.
func (m *Module) GenRemoteStateToSelf() ([]byte, error) {
	return m.BackendPtr.GetRemoteStateHCL(*m)
}

func (m *Module) checkDependMarker(data reflect.Value) (reflect.Value, error) {
	subVal := reflect.ValueOf(data.Interface())
	resString := subVal.String()
	for key, marker := range m.ProjectPtr.DependencyMarkers {
		if strings.Contains(resString, key) {
			if marker.InfraName == "this" {
				marker.InfraName = m.InfraPtr.Name
			}
			modKey := fmt.Sprintf("%s.%s", marker.InfraName, marker.ModuleName)
			depModule, exists := m.ProjectPtr.Modules[modKey]
			if !exists {
				return reflect.ValueOf(nil), fmt.Errorf("Depend module does not exists. Src: '%s.%s', depend: '%s'", m.InfraPtr.Name, m.Name, modKey)
			}
			m.Dependencies = append(m.Dependencies, &Dependency{
				Module: depModule,
				Output: marker.Output,
			})
			depModule.Outputs[marker.Output] = true
			//remoteStateRef := fmt.Sprintf("${data.terraform_remote_state.%s-%s.outputs.%s}", marker.InfraName, marker.ModuleName, marker.Output)
			// log.Debugf("Module: %v\nDep: %v", depModule, remoteStateRef)
			// replacer := strings.NewReplacer(key, remoteStateRef)
			// resString = replacer.Replace(resString)
			return reflect.ValueOf(resString), nil
		}
	}
	return subVal, nil
}

func (m *Module) checkYAMLBlockMarker(data reflect.Value) (reflect.Value, error) {
	subVal := reflect.ValueOf(data.Interface())
	for hash := range m.ProjectPtr.InsertYAMLMarkers {
		if subVal.String() == hash {
			return reflect.ValueOf(m.ProjectPtr.InsertYAMLMarkers[hash]), nil
		}
	}
	return subVal, nil
}

// ReplaceMarkers replace all templated markers with values.
func (m *Module) ReplaceMarkers() error {
	err := processingMarkers(m.Inputs, m.checkYAMLBlockMarker)
	if err != nil {
		return err
	}
	err = processingMarkers(m.Inputs, m.checkDependMarker)
	if err != nil {
		return err
	}
	return nil
}

func (m *Module) generateScripts(subCmd string) (string, error) {

	tfCmd := `
# Module '{{ .module }}' infra '{{ .infra }}'.
pushd {{ .infra }}.{{ .module }}
mkdir .terraform
rm -f ./.terraform/plugins
ln -s  ../../../.tmp/plugins ./.terraform/plugins
terraform init
terraform {{ .command }} -auto-approve
popd

`
	t := map[string]interface{}{
		"module":  m.Name,
		"infra":   m.InfraPtr.Name,
		"command": subCmd,
	}
	tmpl, err := template.New("main").Option("missingkey=error").Parse(tfCmd)

	if err != nil {
		return "", err
	}
	templatedConf := bytes.Buffer{}
	err = tmpl.Execute(&templatedConf, &t)
	if err != nil {
		return "", err
	}

	return templatedConf.String(), nil
}
