package common

import (
	"fmt"
	"io/fs"
	"os"

	"github.com/apex/log"
	"github.com/shalb/cluster.dev/pkg/utils"
)

// CreateCodeDir generate all terraform code for project.
func (m *Unit) createCodeDir() error {
	err := os.Mkdir(m.CacheDir, 0755)
	if err != nil {
		return fmt.Errorf("read unit '%v': mkdir '%v': '%v'", m.Name(), m.CacheDir, err.Error())
	}
	if m.WorkDir != "" {
		err = utils.CopyDirectory(m.WorkDir, m.CacheDir)
		if err != nil {
			return fmt.Errorf("read unit '%v': creating cache: '%v'", m.Name(), err.Error())
		}
	}
	for _, f := range *m.CreateFiles {
		if m.ProjectPtr.CheckContainsMarkers(f.Content) {
			log.Debugf("Unprocessed markers:\n %+v", f.Content)
			return fmt.Errorf("misuse of functions in a template: unit: '%s.%s'", m.StackPtr.Name, m.MyName)
		}
	}

	return m.CreateFiles.WriteFiles(m.CacheDir)
}

func (m *Unit) Build() error {

	if m.PreHook != nil {
		m.CreateFiles.Add("pre_hook.sh", m.PreHook.Command, fs.ModePerm)
	}
	if m.PostHook != nil {
		m.CreateFiles.Add("post_hook.sh", m.PostHook.Command, fs.ModePerm)
	}
	return m.createCodeDir()
}
