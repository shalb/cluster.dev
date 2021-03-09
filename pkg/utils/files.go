package utils

import (
	"os"
	"regexp"
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

func IsDir(path string) (bool, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	if fileInfo.IsDir() {
		return true, nil
	}
	return false, nil
}

// fileInfo, err := os.Stat(absSource)
// if err != nil {
// 	return err
// }
