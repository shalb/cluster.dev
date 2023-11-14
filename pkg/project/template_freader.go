package project

import (
	"os"
	"path/filepath"

	"github.com/apex/log"
)

type tmplFileReader struct {
	stackPtr *Stack
}

func (t tmplFileReader) ReadFile(path string) (string, error) {
	vfPath := filepath.Join(t.stackPtr.TemplateDir, path)
	valuesFileContent, err := os.ReadFile(vfPath)
	if err != nil {
		log.Debugf(err.Error())
		return "", err
	}
	return string(valuesFileContent), nil
}

func (t tmplFileReader) TemplateFile(path string) (string, error) {
	vfPath := filepath.Join(t.stackPtr.TemplateDir, path)
	rawFile, err := os.ReadFile(vfPath)
	if err != nil {
		log.Debugf(err.Error())
		return "", err
	}
	templatedFile, errIsWarn, err := t.stackPtr.TemplateTry(rawFile, vfPath)
	if err != nil {
		if !errIsWarn {
			return "", err
		}
	}
	return string(templatedFile), nil
}
