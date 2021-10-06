package common

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/apex/log"
	"github.com/shalb/cluster.dev/pkg/utils"
)

// CreateCodeDir generate all terraform code for project.
func (m *Unit) CreateCodeDir() error {
	err := os.Mkdir(m.cacheDir, 0755)
	if err != nil {
		return fmt.Errorf("read unit '%v': mkdir '%v': '%v'", m.Name(), m.cacheDir, err.Error())
	}
	err = utils.CopyDirectory(m.WorkDir, m.cacheDir)
	if err != nil {
		return fmt.Errorf("read unit '%v': creating cache: '%v'", m.Name(), err.Error())
	}
	for _, f := range m.CreateFiles {
		filePath := filepath.Join(m.cacheDir, f.File)
		// relPath, _ := filepath.Rel(config.Global.WorkingDir, filePath)
		if m.projectPtr.CheckContainsMarkers(f.Content) {
			log.Debugf("Unprocessed markers:\n %+v", f.Content)
			return fmt.Errorf("misuse of functions in a template: unit: '%s.%s'", m.stackPtr.Name, m.MyName)
		}
		if utils.FileExists(filePath) {
			return fmt.Errorf("read unit '%v': creating cache: file '%v' is already  exists in working dir", m.Name(), f.File)
		}
		err = ioutil.WriteFile(filePath, []byte(f.Content), 0777)
		if err != nil {
			log.Debug(err.Error())
			return err
		}
	}
	return nil
}

func (m *Unit) Build() error {
	return m.CreateCodeDir()
}
