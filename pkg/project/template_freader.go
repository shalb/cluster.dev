package project

import (
	"io/ioutil"
	"path/filepath"

	"github.com/apex/log"
)

type tmplFileReader struct {
	stackPtr *Stack
}

func (t tmplFileReader) ReadFile(path string) (string, error) {
	vfPath := filepath.Join(t.stackPtr.TemplateDir, path)
	valuesFileContent, err := ioutil.ReadFile(vfPath)
	if err != nil {
		log.Debugf(err.Error())
		return "", err
	}
	return string(valuesFileContent), nil
}

func (t tmplFileReader) TemplateFile(path string) (string, error) {
	vfPath := filepath.Join(t.stackPtr.TemplateDir, path)
	rawFile, err := ioutil.ReadFile(vfPath)
	if err != nil {
		log.Debugf(err.Error())
		return "", err
	}
	templatedFile, errIsWarn, err := t.stackPtr.TemplateTry(rawFile, vfPath)
	if err != nil {
		if !errIsWarn {
			log.Fatal(err.Error())
		}
	}
	return string(templatedFile), nil
}
