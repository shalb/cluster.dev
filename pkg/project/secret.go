package project

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"

	"github.com/apex/log"
	"github.com/olekukonko/tablewriter"
	"github.com/shalb/cluster.dev/pkg/config"
	"github.com/shalb/cluster.dev/pkg/executor"
	"github.com/shalb/cluster.dev/pkg/sopstools"
)

const secretObjKindKey = "secret"

type Secret struct {
	filename string
	data     map[string]interface{}
}

func (p *Project) readSecrets() error {
	for filename, data := range config.Global.Manifests {
		isSec, err := FileIsSecret(data, p)
		if err != nil {
			return err
		}
		if !isSec {
			continue
		}
		decryptedRaw, err := sopstools.DecryptYaml(data)
		if err != nil {
			return err
		}
		objects, err := ReadYAMLObjects(decryptedRaw)
		if err != nil {
			log.Debug(err.Error())
			return err
		}
		for _, obj := range objects {
			err = p.readSecretObj(ObjectData{
				filename: filename,
				data:     obj,
			})
			if err != nil {
				log.Debug(err.Error())
				return err
			}
		}
	}
	return nil
}

func (p *Project) readSecretObj(secretSpec ObjectData) error {
	name, ok := secretSpec.data["name"].(string)
	if !ok {
		return fmt.Errorf("Secret object must contain field 'name'")
	}
	// Check if infra with this name is already exists in project.
	if _, ok = p.secrets[name]; ok {
		return fmt.Errorf("Duplicate secrets name '%s'", name)
	}

	p.secrets[name] = Secret{
		filename: secretSpec.filename,
		data:     secretSpec.data,
	}
	if _, exists := p.configData["secrets"]; !exists {
		p.configData["secrets"] = map[string]interface{}{}
	}
	dataForTemplate, ok := secretSpec.data["encrypted_data"]
	if !ok {
		return fmt.Errorf("secret must contain field 'encrypted_data'")
	}
	p.configData["secrets"].(map[string]interface{})[name] = dataForTemplate
	return nil
}

func FileIsSecret(data []byte, p *Project) (bool, error) {
	tmpl, err := template.New("main").Funcs(p.TmplFunctionsMap).Option("missingkey=default").Parse(string(data))
	if err != nil {
		return false, err
	}
	templatedConf := bytes.Buffer{}
	err = tmpl.Execute(&templatedConf, nil)
	if err != nil {
		return false, err
	}
	objects, err := ReadYAMLObjects(templatedConf.Bytes())
	if err != nil {
		log.Debug(err.Error())
		return false, err
	}
	isSec, err := AllObjectsIsSecrets(objects)
	if err != nil {
		log.Debug(err.Error())
		return false, err
	}

	return isSec, nil
}

func AllObjectsIsSecrets(objects []map[string]interface{}) (bool, error) {
	secretsCount := 0
	for _, obj := range objects {
		if kind, ok := obj["kind"].(string); ok {
			if kind == "secret" {
				secretsCount++
			}
		}
	}
	if secretsCount == len(objects) {
		return true, nil
	}
	if secretsCount == 0 {
		return false, nil
	}
	return false, fmt.Errorf("file with secrets must contain only secrets")
}

func (p *Project) PrintSecretsList() {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Name", "File"})
	for name, secret := range p.secrets {
		relPath, _ := filepath.Rel(config.Global.WorkingDir, secret.filename)
		table.Append([]string{name, "./" + relPath})
	}
	table.Render()
}

func (p *Project) EditSecret(name string) error {
	secret, exists := p.secrets[name]
	runner, err := executor.NewBashRunner(config.Global.WorkingDir)
	if err != nil {
		return err
	}
	if exists {
		command := fmt.Sprintf("sops %s", secret.filename)
		err = runner.RunWithTty(command)
		return err
	} else {
		filename, err := createSecretTmpl(name)
		if err != nil {
			return err
		}
		command := fmt.Sprintf("sops -e --encrypted-regex ^encrypted_data$ -i %s", filename)
		err = runner.RunWithTty(command)
		if err != nil {
			return err
		}
		command = fmt.Sprintf("sops %s", filename)
		err = runner.RunWithTty(command)
		if err != nil {
			return err
		}
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
		err = ioutil.WriteFile(filenameCheck, templatedSecret.Bytes(), os.ModePerm)
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
# Only values inside encrypted_data will be encrypted
encrypted_data: 
    key: secret string
    secret_cat:
        int_key: 1
        bool_key: true
    password: PaSworD1
`
