package common

import (
	"fmt"
	"io/fs"
	"os"

	"github.com/shalb/cluster.dev/pkg/project"
	"github.com/shalb/cluster.dev/pkg/utils"
)

// CreateCodeDir generate all terraform code for project.
func (u *Unit) createCodeDir() error {

	err := os.Mkdir(u.CacheDir, 0755)
	if err != nil {
		return fmt.Errorf("build unit '%v': mkdir '%v': '%v'", u.Name(), u.CacheDir, err.Error())
	}
	if u.WorkDir != "" {
		err = utils.CopyDirectory(u.WorkDir, u.CacheDir)
		if err != nil {
			return fmt.Errorf("read unit '%v': creating cache: '%v'", u.Name(), err.Error())
		}
	}
	// for _, f := range *m.CreateFiles {
	// 	if m.ProjectPtr.CheckContainsMarkers(f.Content) {
	// 		log.Debugf("Unprocessed markers:\n %+v", f.Content)
	// 		return fmt.Errorf("misuse of functions in a template: unit: '%s.%s'", m.StackPtr.Name, m.MyName)
	// 	}
	// }

	return u.CreateFiles.WriteFiles(u.CacheDir)
}

func (u *Unit) Build() error {
	// Save state before create code dir.
	u.SavedState = u.GetState()
	err := u.ScanData(project.OutputsReplacer)
	if err != nil {
		return err
	}
	if u.PreHook != nil {
		u.CreateFiles.Add("pre_hook.sh", u.PreHook.Command, fs.ModePerm)
	}
	if u.PostHook != nil {
		u.CreateFiles.Add("post_hook.sh", u.PostHook.Command, fs.ModePerm)
	}

	return u.createCodeDir()
}
