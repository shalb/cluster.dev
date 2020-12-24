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
	depMarkers, ok := m.projectPtr.Markers["DependencyMarkers"]
	if !ok {
		log.Debug("Internal error.")
	}
	for hash, marker := range depMarkers.(map[string]*DependencyMarker) {
		remoteStateRef := fmt.Sprintf("data.terraform_remote_state.%s-%s.outputs.%s", marker.InfraName, marker.ModuleName, marker.Output)
		hcltools.ReplaceStingMarkerInBody(moduleBody, hash, remoteStateRef)
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
	for output := range m.ExpectedOutputs {
		log.Debugf("Output: %v", output)
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
	depsUniq := map[*project.Module]bool{}

	for _, dep := range *m.Dependencies() {
		if _, ok := depsUniq[dep.Module]; ok {
			continue
		}
		depsUniq[dep.Module] = true
		rs, err := m.genRemoteStateToSelf()
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
	modDir := filepath.Join(codeDir, m.Name())
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
terraform {{ .command }} -auto-approve
popd

`
	t := map[string]interface{}{
		"module":  m.Name(),
		"infra":   m.InfraName(),
		"command": subCmd,
	}
	tmpl, err := template.New("main").Option("missingkey=error").Parse(tfCmd)

	if err != nil {
		log.Trace(err.Error())
		return ""
	}
	templatedConf := bytes.Buffer{}
	err = tmpl.Execute(&templatedConf, &t)
	if err != nil {
		log.Trace(err.Error())
		return ""
	}

	return templatedConf.String()
}
