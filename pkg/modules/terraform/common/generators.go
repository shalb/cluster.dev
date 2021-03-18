package common

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/apex/log"
	"github.com/shalb/cluster.dev/pkg/config"
	"github.com/shalb/cluster.dev/pkg/hcltools"
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
func (m *Module) CreateCodeDir(projectCodeDir string) error {

	modDir := filepath.Join(projectCodeDir, m.Key())
	log.Infof("Generating code for module module '%v'", m.Key())
	err := os.Mkdir(modDir, 0755)

	for fn, f := range m.FilesList() {
		filePath := filepath.Join(modDir, fn)
		relPath, _ := filepath.Rel(config.Global.WorkingDir, filePath)
		log.Debugf(" file: './%v'", relPath)
		if m.projectPtr.CheckContainsMarkers(string(f)) {
			log.Debugf("Unprocessed markers:\n %+v", string(f))
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
		m.filesList["pre_hook.sh"] = []byte(m.preHook.Command)
	}
	if m.postHook != nil {
		m.filesList["post_hook.sh"] = []byte(m.postHook.Command)
	}
	return nil
}
