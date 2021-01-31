package common

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/apex/log"
	"github.com/shalb/cluster.dev/pkg/project"
)

// genBackendCodeBlock generate backend code block for this module.
func (m *Module) genBackendCodeBlock() ([]byte, error) {

	res, err := m.backendPtr.GetBackendHCL(m.InfraName(), m.Name())
	if err != nil {
		log.Debug(err.Error())
		return nil, err
	}
	return res, nil
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
		rs, err := modBackend.GetRemoteStateHCL(dep.Module.InfraName(), dep.Module.Name())
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

	modDir := filepath.Join(codeDir, m.Key())
	log.Infof("Generating code for module module '%v'", m.Key())
	err := os.Mkdir(modDir, 0755)

	for fn, f := range m.FilesList {
		filePath := filepath.Join(modDir, fn)
		log.Debugf(" file: '%v'", filePath)
		if m.projectPtr.CheckContainsMarkers(string(f)) {
			log.Debugf("%+v", string(f))
			log.Fatalf("Unprocessed remote marker found in module '%s.%s' (backend block). Check documentation.", m.infraPtr.Name, m.name)
		}
		err = ioutil.WriteFile(filePath, f, 0777)
		if err != nil {
			log.Debug(err.Error())
			return err
		}
	}
	m.codeDir = modDir
	return nil
}

func (m *Module) BuildCommon() error {
	var err error
	m.FilesList["init.tf"], err = m.genBackendCodeBlock()
	if err != nil {
		log.Debug(err.Error())
		return err
	}
	// Create remote_state.tf
	remoteStates, err := m.genDepsRemoteStates()
	if err != nil {
		log.Debug(err.Error())
		return err
	}
	if len(remoteStates) > 0 {
		m.FilesList["remote_states.tf"], err = m.genDepsRemoteStates()
	}

	if m.preHook != nil {
		m.FilesList["pre_hook.sh"] = m.preHook
	}
	if m.postHook != nil {
		m.FilesList["post_hook.sh"] = m.preHook
	}
	return nil
}
