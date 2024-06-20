package utils

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"

	"github.com/apex/log"
)

// FileExists check if file exists.
func FileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

func IsLocalPath(path string) bool {
	re := regexp.MustCompile(`^(/|\./|\.\./).*`)
	outputName := re.FindString(path)
	return len(outputName) >= 1
}

func IsAbsolutePath(path string) bool {
	if len(path) < 1 {
		return false
	}
	if path[0:1] == "/" {
		return true
	}
	return false
}

// CheckDir check if 'pash' exists and is directory.
// error means os.Stat error.
// true - dir
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
		return fmt.Errorf("ReadFilesToExistentsList: %w", err)
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
					return fmt.Errorf("ReadFilesToExistentsList: %w", err)
				}
				filesList[relPath], err = os.ReadFile(path)
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
		err = os.WriteFile(fileFullName, []byte(fData), os.ModePerm)
		if err != nil {
			return err
		}
	}
	return
}

func CopyDirectory(scrDir, dest string) error {
	entries, err := os.ReadDir(scrDir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		sourcePath := filepath.Join(scrDir, entry.Name())
		destPath := filepath.Join(dest, entry.Name())

		fileInfo, err := os.Stat(sourcePath)
		if err != nil {
			return err
		}

		stat, ok := fileInfo.Sys().(*syscall.Stat_t)
		if !ok {
			return fmt.Errorf("failed to get raw syscall.Stat_t data for '%s'", sourcePath)
		}

		switch fileInfo.Mode() & os.ModeType {
		case os.ModeDir:
			if err := CreateIfNotExists(destPath, 0755); err != nil {
				return err
			}
			if err := CopyDirectory(sourcePath, destPath); err != nil {
				return err
			}
		case os.ModeSymlink:
			if err := CopySymLink(sourcePath, destPath); err != nil {
				return err
			}
		default:
			if err := Copy(sourcePath, destPath); err != nil {
				return err
			}
		}

		if err := os.Lchown(destPath, int(stat.Uid), int(stat.Gid)); err != nil {
			return err
		}

		isSymlink := fileInfo.Mode()&os.ModeSymlink != 0
		if !isSymlink {
			if err := os.Chmod(destPath, fileInfo.Mode()); err != nil {
				return err
			}
		}
	}
	return nil
}

func Copy(srcFile, dstFile string) error {
	out, err := os.Create(dstFile)
	if err != nil {
		return err
	}

	defer out.Close()

	in, err := os.Open(srcFile)
	if err != nil {
		return err
	}
	defer in.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}

	return nil
}

func Exists(filePath string) bool {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return false
	}

	return true
}

func CreateIfNotExists(dir string, perm os.FileMode) error {
	if Exists(dir) {
		return nil
	}

	if err := os.MkdirAll(dir, perm); err != nil {
		return fmt.Errorf("failed to create directory: '%s', error: '%s'", dir, err.Error())
	}

	return nil
}

func CopySymLink(source, dest string) error {
	link, err := os.Readlink(source)
	if err != nil {
		return err
	}
	return os.Symlink(link, dest)
}

func RemoveDirContent(dir string) error {
	// log.Warnf("removeDirContent: %v", dir)
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer d.Close()
	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}
	for _, name := range names {
		err = os.RemoveAll(filepath.Join(dir, name))
		if err != nil {
			return err
		}
	}
	return nil
}
