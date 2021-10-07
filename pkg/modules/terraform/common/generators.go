package common

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/apex/log"
	"github.com/shalb/cluster.dev/pkg/hcltools"
	"github.com/shalb/cluster.dev/pkg/project"
)

// genBackendCodeBlock generate backend code block for this unit.
func (m *Unit) genBackendCodeBlock() ([]byte, error) {

	f, err := m.backendPtr.GetBackendHCL(m.StackName(), m.Name())
	if err != nil {
		log.Debug(err.Error())
		return nil, err
	}
	if len(m.requiredProviders) < 1 {
		return f.Bytes(), nil
	}
	tb := f.Body().Blocks()[0]
	tfBlock := tb.Body().AppendNewBlock("required_providers", []string{})
	for name, prov := range m.requiredProviders {

		reqProvs, err := hcltools.InterfaceToCty(prov)
		if err != nil {
			return nil, err
		}
		tfBlock.Body().SetAttributeValue(name, reqProvs)
	}
	return f.Bytes(), nil
}

// genDepsRemoteStates generate terraform remote states for all dependencies of this unit.
func (m *Unit) genDepsRemoteStates() ([]byte, error) {
	var res []byte
	depsUniq := map[project.Unit]bool{}
	for _, dep := range *m.Dependencies() {
		// Ignore duplicated dependencies.
		if _, ok := depsUniq[dep.Unit]; ok {
			continue
		}
		// Ignore dependencies without output (user defined as 'depends_on' option.)
		if dep.Output == "" {
			continue
		}
		// Deduplication.
		depsUniq[dep.Unit] = true
		modBackend := dep.Unit.StackPtr().Backend
		rs, err := modBackend.GetRemoteStateHCL(dep.Unit.StackName(), dep.Unit.Name())
		if err != nil {
			log.Debug(err.Error())
			return nil, err
		}
		res = append(res, rs...)
	}
	return res, nil
}

// CreateCodeDir generate all terraform code for project.
func (m *Unit) CreateCodeDir() error {
	err := os.Mkdir(m.codeDir, 0755)

	for fn, f := range m.FilesList() {
		filePath := filepath.Join(m.codeDir, fn)
		// relPath, _ := filepath.Rel(config.Global.WorkingDir, filePath)
		if m.projectPtr.CheckContainsMarkers(string(f)) {
			log.Debugf("Unprocessed markers:\n %+v", string(f))
			return fmt.Errorf("misuse of functions in a template: unit: '%s.%s'", m.stackPtr.Name, m.name)
		}
		err = ioutil.WriteFile(filePath, f, 0777)
		if err != nil {
			log.Debug(err.Error())
			return err
		}
	}
	return nil
}

func (m *Unit) Build() error {
	var err error

	m.filesList["init.tf"], err = m.genBackendCodeBlock()
	if err != nil {
		log.Debug(err.Error())
		return err
	}
	if m.providers != nil {
		providers, err := hcltools.ProvidersToHCL(m.providers)
		if err != nil {
			log.Debug(err.Error())
			return err
		}
		m.filesList["init.tf"] = append(m.filesList["init.tf"], providers.Bytes()...)
	}

	// Create remote_state.tf
	remoteStates, err := m.genDepsRemoteStates()
	if err != nil {
		log.Debug(err.Error())
		return err
	}
	if len(remoteStates) > 0 {
		m.filesList["remote_states.tf"], err = m.genDepsRemoteStates()
	}
	if err != nil {
		log.Debug(err.Error())
		return err
	}

	if m.preHook != nil {
		hookCmd, err := m.replaceRemoteStatesForBash(m.preHook.Command)
		if err != nil {
			return err
		}
		m.filesList["pre_hook.sh"] = []byte(hookCmd)
	}
	if m.postHook != nil {
		hookCmd, err := m.replaceRemoteStatesForBash(m.postHook.Command)
		if err != nil {
			return err
		}
		m.filesList["post_hook.sh"] = []byte(hookCmd)
	}
	return nil
}

func (m *Unit) replaceRemoteStatesForBash(cmd string) (res string, err error) {
	res = cmd
	for hash, mr := range m.Markers() {
		marker, ok := mr.(*project.DependencyOutput)
		if !ok {
			return "", fmt.Errorf("preparing unit: internal error: incorrect remote state type")
		}
		refStr := DependencyToBashRemoteState(marker)
		res = strings.ReplaceAll(res, hash, refStr)
	}
	return
}
