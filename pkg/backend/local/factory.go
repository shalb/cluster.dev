package local

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/apex/log"
	"github.com/shalb/cluster.dev/pkg/config"
	"github.com/shalb/cluster.dev/pkg/project"
	"github.com/shalb/cluster.dev/pkg/utils"
	"gopkg.in/yaml.v3"
)

// Factory factory for local backends.
type Factory struct{}

// New creates the new local backend.
func (f *Factory) New(cnf []byte, name string, p *project.Project) (project.Backend, error) {
	bk := Backend{
		name:       name,
		ProjectPtr: p,
	}
	err := yaml.Unmarshal(cnf, &bk)
	if err != nil {
		return nil, utils.ResolveYamlError(cnf, err)
	}
	defaultBackendPath := fmt.Sprintf("%s/states", config.Global.WorkDir)
	if bk.Path == "" {
		bk.Path = defaultBackendPath
	}
	if !utils.IsAbsolutePath(bk.Path) {
		bk.Path = filepath.Join(config.Global.ProjectConfigsPath, bk.Path)
	}
	isDir, err := utils.CheckDir(bk.Path)
	if isDir {
		return &bk, nil
	}

	log.Debugf("Creating local backend dir: %v", bk.Path)
	err = os.MkdirAll(bk.Path, os.ModePerm)
	return &bk, err
}

func init() {
	log.Debug("Registering backend provider local..")
	if err := project.RegisterBackendFactory(&Factory{}, "local"); err != nil {
		log.Trace("Can't register backend provider local.")
	}
}
