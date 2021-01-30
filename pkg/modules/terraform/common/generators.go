package common

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

// genMainCodeBlockHCL generate main code block for this module.
func (m *Module) genMainCodeBlockHCL() ([]byte, error) {
	return nil, nil
}

// genBackendCodeBlock generate backend code block for this module.
func (m *Module) genBackendCodeBlock() ([]byte, error) {

	res, err := m.backendPtr.GetBackendHCL(m)
	if err != nil {
		log.Debug(err.Error())
		return nil, err
	}
	return res, nil
}

// genOutputsBlock generate output code block for this module.
func (m *Module) genOutputsBlock() ([]byte, error) {
	return nil, nil
}

// genDepsRemoteStates generate terraform remote states for all dependencies of this module.
func (m *Module) genDepsRemoteStates() ([]byte, error) {
	var res []byte
	depsUniq := map[project.Module]bool{}
	for _, dep := range *m.Dependencies() {
		// Ignore duplicated dependencies.
		if _, ok := depsUniq[dep.Module]; ok {
			continue
		}
		// Ignore dependencies without output (user defined as 'depends_on' option.)
		if dep.Output == "" {
			continue
		}
		// Deduplication.
		depsUniq[dep.Module] = true
		modBackend := dep.Module.InfraPtr().Backend
		rs, err := modBackend.GetRemoteStateHCL(dep.Module)
		if err != nil {
			log.Debug(err.Error())
			return nil, err
		}
		res = append(res, rs...)
	}
	return res, nil
}

// CreateCodeDir generate all terraform code for project.
func (m *Module) CreateCodeDir(codeDir string) error {

	mName := fmt.Sprintf("%s.%s", m.InfraName(), m.Name())
	modDir := filepath.Join(codeDir, mName)
	log.Infof("Generating code for module module '%v'", mName)
	err := os.Mkdir(modDir, 0755)
	// if err != nil {
	// 	return err
	// }
	// Create main.tf

	tfFile := filepath.Join(modDir, "main.tf")
	log.Debugf(" file: '%v'", tfFile)
	codeBlock, err := m.genMainCodeBlockHCL()
	if err != nil {
		log.Debug(err.Error())
		return err
	}
	if m.projectPtr.CheckContainsMarkers(string(codeBlock)) {
		log.Fatalf("Unprocessed remote state pointer found in module '%s.%s' (module block). Template function remoteState can only be used as a yaml value or a part of yaml value.", m.infraPtr.Name, m.name)
	}
	ioutil.WriteFile(tfFile, codeBlock, os.ModePerm)
	if err != nil {
		log.Debug(err.Error())
		return err
	}
	// Create init.tf
	tfFile = filepath.Join(modDir, "init.tf")
	log.Debugf(" file: '%v'", tfFile)
	codeBlock, err = m.genBackendCodeBlock()
	if err != nil {
		log.Debug(err.Error())
		return err
	}
	if m.projectPtr.CheckContainsMarkers(string(codeBlock)) {
		log.Fatalf("Unprocessed remote state pointer found in module '%s.%s' (backend block). Template function remoteState can only be used as a yaml value or a part of yaml value.", m.infraPtr.Name, m.name)
	}
	ioutil.WriteFile(tfFile, codeBlock, os.ModePerm)
	if err != nil {
		log.Debug(err.Error())
		return err
	}
	// Create remote_state.tf
	codeBlock, err = m.genDepsRemoteStates()
	if err != nil {
		log.Debug(err.Error())
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
			log.Debug(err.Error())
			return err
		}
	}
	// Create outputs.tf
	codeBlock, err = m.genOutputsBlock()
	if err != nil {
		log.Debug(err.Error())
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
			log.Debug(err.Error())
			return err
		}
	}

	if m.preHook != nil {
		preHookFile := filepath.Join(modDir, "pre_hook.sh")
		if m.projectPtr.CheckContainsMarkers(string(m.preHook)) {
			log.Fatalf("Unprocessed remote state pointer found in module '%s.%s' (pre_hook). Template function remoteState and insertYAML can't be used in pre_hook.", m.infraPtr.Name, m.name)
		}
		log.Debugf(" file: '%v'", preHookFile)
		ioutil.WriteFile(preHookFile, m.preHook, 0777)
		if err != nil {
			log.Debug(err.Error())
			return err
		}
	}
	if m.postHook != nil {
		postHookFile := filepath.Join(modDir, "post_hook.sh")
		if m.projectPtr.CheckContainsMarkers(string(m.preHook)) {
			log.Fatalf("Unprocessed remote state pointer found in module '%s.%s' (post_hook). Template function remoteState and insertYAML can't be used in post_hook.", m.infraPtr.Name, m.name)
		}
		log.Debugf(" file: '%v'", postHookFile)
		ioutil.WriteFile(postHookFile, m.postHook, 0777)
		if err != nil {
			log.Debug(err.Error())
			return err
		}
	}
	m.codeDir = modDir
	return nil
}

// GetApplyShellCmd return string with bash commands sequence witch need to run in working dir to apply this module.
func (m *Module) GetApplyShellCmd() string {
	return m.getShellCmd("apply")
}

// GetDestroyShellCmd return string with bash commands sequence witch need to run in working dir to destroy this module.
func (m *Module) GetDestroyShellCmd() string {
	return m.getShellCmd("destroy")
}

func (m *Module) getShellCmd(subCmd string) string {

	tfCmd := `
# Module '{{ .module }}' infra '{{ .infra }}'.
pushd {{ .infra }}.{{ .module }}
mkdir -p .terraform
rm -rf ./.terraform/plugins
mkdir -p ../../../.tmp/plugins
ln -s  ../../../.tmp/plugins ./.terraform/plugins
terraform init
{{- if .pre_hook }}
./pre_hook.sh{{ end }}
terraform {{ .command }} -auto-approve
{{- if .post_hook }}
./post_hook.sh{{ end }}
{{ .outputs }}
popd

`
	var outputsShellBlock string
	var postHook, preHook bool
	preHook = m.preHook != nil
	postHook = m.postHook != nil
	t := map[string]interface{}{
		"module":    m.Name(),
		"infra":     m.InfraName(),
		"command":   subCmd,
		"outputs":   outputsShellBlock,
		"pre_hook":  preHook,
		"post_hook": postHook,
	}
	tmpl, err := template.New("main").Option("missingkey=error").Parse(tfCmd)

	if err != nil {
		log.Trace(err.Error())
		return ""
	}
	templatedConf := bytes.Buffer{}
	err = tmpl.Execute(&templatedConf, &t)
	if err != nil {
		log.Error(err.Error())
		return ""
	}

	return templatedConf.String()
}
