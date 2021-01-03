package terraform

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"text/template"

	"github.com/apex/log"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/shalb/cluster.dev/pkg/hcltools"
	"github.com/shalb/cluster.dev/pkg/project"
	"github.com/zclconf/go-cty/cty"
)

// genMainCodeBlockHCL generate main code block for this module.
func (m *TFModule) genMainCodeBlockHCL() ([]byte, error) {
	f := hclwrite.NewEmptyFile()
	rootBody := f.Body()
	moduleBlock := rootBody.AppendNewBlock("module", []string{m.name})
	moduleBody := moduleBlock.Body()
	moduleBody.SetAttributeValue("source", cty.StringVal(m.Source))
	for key, val := range m.Inputs {
		ctyVal, err := hcltools.InterfaceToCty(val)
		if err != nil {
			return nil, err
		}
		moduleBody.SetAttributeValue(key, ctyVal)
	}
	depMarkers, ok := m.projectPtr.Markers[remoteStateMarkerCatName]
	if !ok {
		log.Debug("Internal error.")
		return nil, fmt.Errorf("internal error")
	}
	for hash, marker := range depMarkers.(map[string]*project.Dependency) {
		remoteStateRef := fmt.Sprintf("data.terraform_remote_state.%s-%s.outputs.%s", marker.Module.InfraName(), marker.Module.Name(), marker.Output)
		hcltools.ReplaceStingMarkerInBody(moduleBody, hash, remoteStateRef)
	}
	outputMarkers, ok := m.projectPtr.Markers[project.OutputMarkerCatName]
	if !ok {
		log.Debug("Internal error.")
		return nil, fmt.Errorf("internal error")
	}
	for hash, marker := range outputMarkers.(map[string]*project.Dependency) {
		vName := project.ConvertToTfVarName(fmt.Sprintf("%s_%s_%s", marker.Module.InfraName(), marker.Module.Name(), marker.Output))
		outputRef := fmt.Sprintf("var.%s", vName)
		hcltools.ReplaceStingMarkerInBody(moduleBody, hash, outputRef)
	}

	for _, d := range m.dependenciesOutputs {
		if d.Output == "" {
			continue
		}
		vName := project.ConvertToTfVarName(fmt.Sprintf("%s_%s_%s", d.Module.InfraName(), d.Module.Name(), d.Output))
		vBlock := rootBody.AppendNewBlock("variable", []string{vName})
		vBlockBody := vBlock.Body()
		vBlockBody.SetAttributeValue("type", cty.StringVal("string"))
	}
	return f.Bytes(), nil
}

// genBackendCodeBlock generate backend code block for this module.
func (m *TFModule) genBackendCodeBlock() ([]byte, error) {

	res, err := m.BackendPtr.GetBackendHCL(m)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// genOutputsBlock generate output code block for this module.
func (m *TFModule) genOutputsBlock() ([]byte, error) {
	f := hclwrite.NewEmptyFile()

	rootBody := f.Body()
	exOutputs := map[string]bool{}

	for output := range m.expectedRemoteStates {
		exOutputs[output] = true
	}
	for output := range m.expectedOutputs {
		exOutputs[output] = true
	}

	for output := range exOutputs {
		re := regexp.MustCompile(`^[A-Za-z][a-zA-Z0-9_\-]{0,}`)
		outputName := re.FindString(output)
		if len(outputName) < 1 {
			return nil, fmt.Errorf("invalid output '%v' in module '%v'", output, m.Name())
		}
		dataBlock := rootBody.AppendNewBlock("output", []string{outputName})
		dataBody := dataBlock.Body()
		outputStr := fmt.Sprintf("module.%s.%s", m.Name(), outputName)
		dataBody.SetAttributeRaw("value", hcltools.CreateTokensForOutput(outputStr))
	}
	return f.Bytes(), nil

}

// genDepsRemoteStates generate terraform remote states for all dependencies of this module.
func (m *TFModule) genDepsRemoteStates() ([]byte, error) {
	var res []byte
	depsUniq := map[project.Module]bool{}
	for _, dep := range m.Dependencies() {
		if _, ok := depsUniq[dep.Module]; ok {
			continue
		}
		depsUniq[dep.Module] = true
		convertedMod, ok := dep.Module.Self().(*TFModule)
		if !ok {
			continue
		}
		rs, err := convertedMod.genRemoteStateToSelf()
		if err != nil {
			return nil, err
		}
		res = append(res, rs...)
	}
	return res, nil
}

// genRemoteStateToSelf - remote state block generate terraform code. It's remote state !to this module! witch should be used in another module depend of this.
func (m *TFModule) genRemoteStateToSelf() ([]byte, error) {
	return m.BackendPtr.GetRemoteStateHCL(m)
}

// CreateCodeDir generate all terraform code for project.
func (m *TFModule) CreateCodeDir(codeDir string) error {

	mName := fmt.Sprintf("%s.%s", m.InfraName(), m.Name())
	modDir := filepath.Join(codeDir, mName)
	log.Infof("Generating code for module module '%v'", mName)
	err := os.Mkdir(modDir, 0755)
	if err != nil {
		return err
	}
	// Create main.tf
	tfFile := filepath.Join(modDir, "main.tf")
	log.Debugf(" file: '%v'", tfFile)
	codeBlock, err := m.genMainCodeBlockHCL()
	if err != nil {
		log.Fatal(err.Error())
		return err
	}
	if m.projectPtr.CheckContainsMarkers(string(codeBlock)) {
		log.Fatalf("Unprocessed remote state pointer found in module '%s.%s' (module block). Template function remoteState can only be used as a yaml value or a part of yaml value.", m.infraPtr.Name, m.name)
	}
	ioutil.WriteFile(tfFile, codeBlock, os.ModePerm)
	if err != nil {
		return err
	}
	// Create init.tf
	tfFile = filepath.Join(modDir, "init.tf")
	log.Debugf(" file: '%v'", tfFile)
	codeBlock, err = m.genBackendCodeBlock()
	if err != nil {
		return err
	}
	if m.projectPtr.CheckContainsMarkers(string(codeBlock)) {
		log.Fatalf("Unprocessed remote state pointer found in module '%s.%s' (backend block). Template function remoteState can only be used as a yaml value or a part of yaml value.", m.infraPtr.Name, m.name)
	}
	ioutil.WriteFile(tfFile, codeBlock, os.ModePerm)
	if err != nil {
		return err
	}
	// Create remote_state.tf
	codeBlock, err = m.genDepsRemoteStates()
	if err != nil {
		return err
	}
	if len(codeBlock) > 1 {
		tfFile = filepath.Join(modDir, "remote_state.tf")
		if m.projectPtr.CheckContainsMarkers(string(codeBlock)) {
			log.Fatalf("Unprocessed remote state pointer found in module '%s.%s' (remote states block). Template function remoteState can only be used as a yaml value or a part of yaml value.", m.infraPtr.Name, m.name)
		}
		log.Debugf(" file: '%v'", tfFile)
		ioutil.WriteFile(tfFile, codeBlock, os.ModePerm)
		if err != nil {
			return err
		}
	}
	// Create outputs.tf
	codeBlock, err = m.genOutputsBlock()
	if err != nil {
		return err
	}
	if len(codeBlock) > 1 {
		tfFile = filepath.Join(modDir, "outputs.tf")
		if m.projectPtr.CheckContainsMarkers(string(codeBlock)) {
			log.Fatalf("Unprocessed remote state pointer found in module '%s.%s' (output block). Template function remoteState can only be used as a yaml value or a part of yaml value.", m.infraPtr.Name, m.name)
		}
		log.Debugf(" file: '%v'", tfFile)
		ioutil.WriteFile(tfFile, codeBlock, os.ModePerm)
		if err != nil {
			return err
		}
	}

	return nil
}

// GetApplyShellCmd return string with bash commands sequence witch need to run in working dir to apply this module.
func (m *TFModule) GetApplyShellCmd() string {
	return m.getShellCmd("apply")
}

// GetDestroyShellCmd return string with bash commands sequence witch need to run in working dir to destroy this module.
func (m *TFModule) GetDestroyShellCmd() string {
	return m.getShellCmd("destroy")
}

func (m *TFModule) getShellCmd(subCmd string) string {

	tfCmd := `
# Module '{{ .module }}' infra '{{ .infra }}'.
pushd {{ .infra }}.{{ .module }}
mkdir .terraform
rm -f ./.terraform/plugins
mkdir -p ../../../.tmp/plugins
ln -s  ../../../.tmp/plugins ./.terraform/plugins
terraform init
terraform {{ .command }} {{ .vars }} -auto-approve
{{ .outputs }}
popd

`
	var outputsShellBlock string

	for output := range m.expectedOutputs {
		envVarName := project.ConvertToShellVarName(fmt.Sprintf("%v.%v.%v", m.InfraName(), m.Name(), output))
		outputsShellBlock += fmt.Sprintf("%[1]s=\"$(terraform output %[2]s)\"\nexport %[1]s\n", envVarName, output)
		// log.Debugf("%v", outputsShellBlock)
	}

	var varString string

	for _, v := range m.dependenciesOutputs {
		if len(v.Output) == 0 {
			continue
		}
		envVarName := project.ConvertToShellVar(fmt.Sprintf("%v.%v.%v", v.InfraName, v.ModuleName, v.Output))
		varName := project.ConvertToTfVarName(fmt.Sprintf("%v.%v.%v", v.InfraName, v.ModuleName, v.Output))
		varString += fmt.Sprintf("-var \"%v=%v\" ", varName, envVarName)
	}

	t := map[string]interface{}{
		"module":  m.Name(),
		"infra":   m.InfraName(),
		"command": subCmd,
		"outputs": outputsShellBlock,
		"vars":    varString,
	}
	tmpl, err := template.New("main").Option("missingkey=error").Parse(tfCmd)

	if err != nil {
		log.Trace(err.Error())
		return ""
	}
	templatedConf := bytes.Buffer{}
	err = tmpl.Execute(&templatedConf, &t)
	if err != nil {
		log.Fatal(err.Error())
		return ""
	}

	return templatedConf.String()
}
