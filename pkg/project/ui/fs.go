package ui

import (
	"io/fs"
	"os"
	"path/filepath"
)

type TmplFS interface {
	ReadDir(name string) ([]fs.DirEntry, error)
	ReadFile(name string) ([]byte, error)
}

type TmplFSLocal struct {
	Dir string
}

func NewTmplFS(path string) TmplFS {
	r := TmplFSLocal{
		Dir: path,
	}
	return &r
}

func (t *TmplFSLocal) ReadDir(path string) ([]fs.DirEntry, error) {
	// log.Debugf("ReadDir: %v, %v", t.Dir, path)
	f := os.DirFS(t.Dir)
	return fs.ReadDir(f, filepath.Join(".", path))
}
func (t *TmplFSLocal) ReadFile(path string) ([]byte, error) {
	// log.Debugf("ReadFile: %v, %v", t.Dir, path)
	f := os.DirFS(t.Dir)
	return fs.ReadFile(f, filepath.Join(".", path))
}
