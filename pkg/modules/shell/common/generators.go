package common

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/apex/log"
)

// CreateCodeDir generate all terraform code for project.
func (m *Module) CreateCodeDir() error {
	err := os.Mkdir(m.codeDir, 0755)

	for fn, f := range m.FilesList() {
		filePath := filepath.Join(m.codeDir, fn)
		// relPath, _ := filepath.Rel(config.Global.WorkingDir, filePath)
		if m.projectPtr.CheckContainsMarkers(string(f)) {
			log.Debugf("Unprocessed markers:\n %+v", string(f))
			return fmt.Errorf("misuse of functions in a template: module: '%s.%s'", m.infraPtr.Name, m.name)
		}
		err = ioutil.WriteFile(filePath, f, 0777)
		if err != nil {
			log.Debug(err.Error())
			return err
		}
	}
	return nil
}

func (m *Module) Build() error {

	return nil
}
