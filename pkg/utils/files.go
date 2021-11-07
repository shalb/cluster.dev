package utils

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/apex/log"
)

// FileExists check if file exists.
func FileExists(filename string) bool {
	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return true
}

func IsLocalPath(path string) bool {
	re := regexp.MustCompile(`^(/|\./|\.\./).*`)
	outputName := re.FindString(path)
	if len(outputName) < 1 {
		return false
	}
	return true
}

func IsAbsolutePath(path string) bool {
	if path[1:2] == "/" {
		return true
	}
	return false
}

func CheckDir(path string) (bool, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	if fileInfo.IsDir() {
		return true, nil
	}
	return false, nil
}

func IsDir(path string) bool {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false
	}
	if fileInfo.IsDir() {
		return true
	}
	return false
}

func ReadFilesToList(filesPath, baseDir string) (filesList map[string][]byte, err error) {
	filesList = make(map[string][]byte)
	err = ReadFilesToExistentsList(filesPath, baseDir, filesList)
	return
}

func ReadFilesToExistentsList(filesPath, baseDir string, filesList map[string][]byte) (err error) {
	_, err = filepath.Rel(baseDir, filesPath)
	if err != nil {
		return
	}
	err = filepath.Walk(filesPath,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			// fmt.Println(path, info.Size(), info.Name())
			if !info.IsDir() {
				if _, exists := filesList[path]; exists {
					return fmt.Errorf("the file '%v' already exists in the list", path)
				}
				relPath, err := filepath.Rel(baseDir, path)
				if err != nil {
					return err
				}
				filesList[relPath], err = ioutil.ReadFile(path)
				if err != nil {
					return err
				}
			}
			return nil
		})
	return
}

func WriteFilesFromList(path string, filesList map[string]string) (err error) {
	for fPath, fData := range filesList {
		var fileName, fileDir, fileFullName string
		splittedPath := strings.Split(fPath, "/")
		if len(splittedPath) < 2 {
			fileDir = path
			fileName = fPath
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
		err = ioutil.WriteFile(fileFullName, []byte(fData), os.ModePerm)
		if err != nil {
			return err
		}
	}
	return
}
