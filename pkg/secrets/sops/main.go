package sops

import (
	"bytes"
	"fmt"
	"html/template"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/apex/log"
	"github.com/shalb/cluster.dev/pkg/config"
	"github.com/shalb/cluster.dev/pkg/executor"
	"github.com/shalb/cluster.dev/pkg/project"
	"github.com/shalb/cluster.dev/pkg/sopstools"
)

const sopsKey = "sops"

type sopsDriver struct{}

// Read([]byte) (string, map[string]interface{}, error)
func (s *sopsDriver) Read(rawData []byte) (name string, data interface{}, err error) {
	decryptedRaw, err := sopstools.DecryptYaml(rawData)
	if err != nil {
		err = fmt.Errorf("decrypting sops secret: %v", err.Error())
		return
	}
	secretSpec, err := project.ReadYAML(decryptedRaw)
	if err != nil {
		err = fmt.Errorf("sops: secret must contain string field 'name'")
		return
	}
	name, ok := secretSpec["name"].(string)
	if !ok {
		err = fmt.Errorf("sops: secret must contain string field 'name'")
		return
	}
	data, ok = secretSpec["encrypted_data"].(map[string]interface{})
	if !ok {
		err = fmt.Errorf("sops secret must contain field 'encrypted_data'")
		return
	}

	return
}

func (s *sopsDriver) Key() string {
	return sopsKey
}

func init() {
	err := project.RegisterSecretDriver(&sopsDriver{}, sopsKey)
	if err != nil {
		log.Fatal(err.Error())
	}
}

func (s *sopsDriver) Edit(sec project.Secret) error {
	runner, err := executor.NewBashRunner(config.Global.WorkingDir)
	if err != nil {
		return err
	}
	command := fmt.Sprintf("sops %s", sec.Filename)
	err = runner.RunWithTty(command)
	if err != nil && err.Error() != "exit status 200" {
		log.Debugf("err %+v", err)
		return err
	}
	return nil
}

func (s *sopsDriver) Create(name string) error {
	runner, err := executor.NewBashRunner(config.Global.WorkingDir)
	if err != nil {
		return err
	}

	filename, err := createSecretTmpl(name)
	if err != nil {
		return err
	}
	command := fmt.Sprintf("sops -e --encrypted-regex ^encrypted_data$ -i %s", filename)
	err = runner.RunWithTty(command)
	if err != nil {
		os.RemoveAll(filename)
		return err
	}
	command = fmt.Sprintf("sops %s", filename)
	err = runner.RunWithTty(command)
	if err != nil && err.Error() != "exit status 200" {
		os.RemoveAll(filename)
		log.Debugf("err %+v", err)
		return err
	}

	return nil
}

func createSecretTmpl(name string) (string, error) {
	tmplData := map[string]string{
		"name": name,
	}
	tmpl, err := template.New("main").Option("missingkey=error").Parse(secretTemplate)
	if err != nil {
		return "", err
	}
	templatedSecret := bytes.Buffer{}
	err = tmpl.Execute(&templatedSecret, tmplData)
	if err != nil {
		return "", err
	}
	filenameCheck := filepath.Join(config.Global.WorkingDir, name+".yaml")
	if _, err := os.Stat(filenameCheck); os.IsNotExist(err) {
		err = ioutil.WriteFile(filenameCheck, templatedSecret.Bytes(), fs.ModePerm)
		if err != nil {
			return "", err
		}
		return filenameCheck, nil
	}
	f, err := ioutil.TempFile(config.Global.WorkingDir, name+"_*.yaml")
	if err != nil {
		return "", err
	}
	defer f.Close()
	_, err = f.Write(templatedSecret.Bytes())
	if err != nil {
		return "", err
	}
	return f.Name(), nil
}

var secretTemplate = `
name: {{ .name }}
kind: secret
driver: sops
# Only values inside encrypted_data will be encrypted
encrypted_data: 
    key: secret string
    secret_cat:
        int_key: 1
        bool_key: true
    password: PaSworD1
`
