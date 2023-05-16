package common

import (
	"fmt"
	"io/fs"

	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/apex/log"
)

// CreateFileRepresentation describes the unit's file that will be saved in the unit's working directory when building.
type CreateFileRepresentation struct {
	FileName string      `yaml:"file"`
	FileMode fs.FileMode `yaml:"file_mode,omitempty"`
	Content  string      `yaml:"content"`
}

// FilesListT describes all unit's files will be write to the unit's working directory when building.
type FilesListT []*CreateFileRepresentation

func (l *FilesListT) Len() int {
	return len(*l)
}

func (l *FilesListT) IsEmpty() bool {
	return len(*l) == 0
}

// Find searchs file and returns a pointer to it or nil if not found.
func (l *FilesListT) Find(fileName string) int {
	for i, f := range *l {
		if f.FileName == fileName {
			return i
		}
	}
	return -1
}

func remove(slice []int, s int) []int {
	return append(slice[:s], slice[s+1:]...)
}

func (l *FilesListT) SPrintLs() string {
	var res string
	for _, f := range *l {
		res += fmt.Sprintf("%v: %v\n", f.FileName, f.FileMode)
	}
	return res
}

// Add insert the new file with name fileName, returns error if file with this name already exists.
func (l *FilesListT) Add(fileName string, content string, mode fs.FileMode) error {
	if l.Find(fileName) >= 0 {
		return fmt.Errorf("add file: file '%v' already exists", fileName)
	}
	*l = append(*l,
		&CreateFileRepresentation{
			FileName: fileName,
			Content:  content,
			FileMode: mode,
		})
	return nil
}

// ReadDir recursively reads files in path, saving relative pathes from baseDir.
func (l *FilesListT) ReadDir(path, baseDir string, pattern ...string) (err error) {
	_, err = filepath.Rel(baseDir, path)
	if err != nil {
		return fmt.Errorf("shell unit: ReadDir: %w", err)
	}
	err = filepath.Walk(path,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				relPath, err := filepath.Rel(baseDir, path)
				if err != nil {
					return fmt.Errorf("shell unit: ReadDir: %w", err)
				}
				matchPattern := true
				for _, p := range pattern {
					matchPattern = regexp.MustCompile(p).MatchString(relPath)
					if matchPattern != true {
						break
					}
				}
				if !matchPattern {
					return nil
				}
				content, err := os.ReadFile(path)
				if err != nil {
					return err
				}
				err = l.Add(relPath, string(content), info.Mode())
				if err != nil {
					return err
				}
			}
			return nil
		})
	return
}

// ReadFile reads file to list.
func (l *FilesListT) ReadFile(path, baseDir string) (err error) {
	relPath, err := filepath.Rel(baseDir, path)
	if err != nil {
		return fmt.Errorf("shell unit: ReadFile: %w", err)
	}
	fileInfo, err := os.Stat(path)
	if err != nil {
		return
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return
	}
	err = l.Add(relPath, string(content), fileInfo.Mode())
	if err != nil {
		return
	}
	return
}

// WriteFiles write all files to path.
func (l *FilesListT) WriteFiles(path string) (err error) {
	for _, file := range *l {
		var fileName, fileDir, fileFullName string
		splittedPath := strings.Split(file.FileName, "/")
		if len(splittedPath) < 2 {
			fileDir = path
			fileName = file.FileName
		} else {
			fileDir = filepath.Join(path, filepath.Join(splittedPath[0:len(splittedPath)-1]...))
			fileName = splittedPath[len(splittedPath)-1]
			err = os.MkdirAll(fileDir, os.ModePerm)
			if err != nil {
				return err
			}
		}

		fileFullName = filepath.Join(fileDir, fileName)
		log.Debugf("Writing file: %v", fileFullName)
		err = os.WriteFile(fileFullName, []byte(file.Content), file.FileMode)
		if err != nil {
			return err
		}
	}
	return
}

// Delete delete file with name fileName, do nothing if not found.
func (l *FilesListT) Delete(fileName string) {
	if i := l.Find(fileName); i >= 0 {
		*l = append((*l)[:i], (*l)[i+1:]...)
	}
	return
}
