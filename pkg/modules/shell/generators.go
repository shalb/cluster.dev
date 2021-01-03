package shell

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"

	"github.com/apex/log"
	"github.com/shalb/cluster.dev/pkg/project"
)

// CreateCodeDir generate all code for module.
func (m *Module) CreateCodeDir(codeDir string) error {
	mName := fmt.Sprintf("%s.%s", m.InfraName(), m.Name())
	modDir := filepath.Join(codeDir, mName)
	log.Infof("Generating code for module module '%v'", mName)
	err := os.Mkdir(modDir, 0755)
	if err != nil {
		return err
	}
	scriptFile := filepath.Join(modDir, "exec.sh")
	err = ioutil.WriteFile(scriptFile, []byte(m.scriptData), 0777)
	return nil
}

// GetApplyShellCmd return string with bash commands sequence witch need to run in working dir to apply this module.
func (m *Module) GetApplyShellCmd() string {
	tfCmd := `
# Module '{{ .module }}' infra '{{ .infra }}'.
pushd {{ .infra }}.{{ .module }}
scriptOutputs="$(./exec.sh {{ .args }})"
{{ .outputsBlock }}
popd

`
	outputsExportBlock := ""
	for eOutput := range m.expectedOutputs {
		oName := fmt.Sprintf("%v.%v.%v", m.InfraName(), m.Name(), eOutput)
		outputsExportBlock += fmt.Sprintf("%[2]v=$(echo ${scriptOutputs} | grep \"set_output_%[1]v=\" | awk -F \"set_output_%[1]v=\" '{print $2}')\n", eOutput, project.ConvertToShellVarName(oName))
		outputsExportBlock += fmt.Sprintf("export %v\n", project.ConvertToShellVarName(oName))
	}
	//  set_output_kubeconfig_path=../kubeconfig{{ .name }}
	args := ""
	for _, arg := range m.Inputs {
		args = fmt.Sprintf("%s %s", args, arg)
	}
	t := map[string]interface{}{
		"module":       m.Name(),
		"infra":        m.InfraName(),
		"args":         args,
		"outputsBlock": outputsExportBlock,
	}
	tmpl, err := template.New("main").Option("missingkey=error").Parse(tfCmd)

	if err != nil {
		log.Fatal(err.Error())
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

// GetDestroyShellCmd return string with bash commands sequence witch need to run in working dir to destroy this module.
func (m *Module) GetDestroyShellCmd() string {
	return m.GetApplyShellCmd()
}
