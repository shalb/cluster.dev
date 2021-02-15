package ssm

import (
	"bytes"
	"fmt"
	"html/template"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/apex/log"
	"github.com/shalb/cluster.dev/pkg/aws"
	"github.com/shalb/cluster.dev/pkg/config"
	"github.com/shalb/cluster.dev/pkg/executor"
	"github.com/shalb/cluster.dev/pkg/project"
	"gopkg.in/yaml.v3"
)

const ssmKey = "aws_ssm"

type ssmDriver struct{}

type ssmSpec struct {
	Region     string `yaml:"region"`
	SecretName string `yaml:"ssm_secret"`
}

// Read([]byte) (string, map[string]interface{}, error)
func (s *ssmDriver) Read(rawData []byte) (name string, data interface{}, err error) {
	secretSpec, err := project.ReadYAML(rawData)
	if err != nil {
		return
	}
	name, ok := secretSpec["name"].(string)
	if !ok {
		err = fmt.Errorf("aws_ssm: secret must contain string field 'name'")
		return
	}
	sp, ok := secretSpec["spec"].(map[string]interface{})
	if !ok {
		err = fmt.Errorf("aws_ssm: secret must contain field 'region'")
		return
	}
	specRaw, err := yaml.Marshal(sp)
	if err != nil {
		err = fmt.Errorf("aws_ssm: can't parse secret '%v' spec %v", name, err)
		return
	}
	var spec ssmSpec
	err = yaml.Unmarshal(specRaw, &spec)
	if err != nil {
		err = fmt.Errorf("aws_ssm: can't parse secret '%v' spec %v", name, err)
		return
	}
	if spec.Region == "" || spec.SecretName == "" {
		if err != nil {
			err = fmt.Errorf("aws_ssm: can't parse secret '%v', fields 'spec.region' and 'spec.secret_name' are required", name)
			return
		}
	}
	data, err = aws.GetSecret(spec.Region, spec.SecretName)
	return
}

func (s *ssmDriver) Key() string {
	return ssmKey
}

func init() {
	err := project.RegisterSecretDriver(&ssmDriver{}, ssmKey)
	if err != nil {
		log.Fatalf("secrets: ssm driver init: %v", err.Error())
	}
}

func (s *ssmDriver) Edit(sec project.Secret) error {
	runner, err := executor.NewBashRunner(config.Global.WorkingDir)
	if err != nil {
		return err
	}
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi"
	}
	command := fmt.Sprintf("%s %s", editor, sec.Filename)
	err = runner.RunWithTty(command)
	if err != nil {
		return err
	}
	return nil
}

func (s *ssmDriver) Create(name string) error {
	runner, err := executor.NewBashRunner(config.Global.WorkingDir)
	if err != nil {
		return fmt.Errorf("secrets: create secret: %v", err.Error())
	}

	filename, err := createSecretTmpl(name)
	if err != nil {
		return fmt.Errorf("secrets: %v", err.Error())
	}
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi"
	}
	command := fmt.Sprintf("%s %s", editor, filename)
	err = runner.RunWithTty(command)
	if err != nil {
		os.RemoveAll(filename)
		log.Debugf("err %+v", err)
		return fmt.Errorf("secrets: create secret: %v", err.Error())
	}

	return nil
}

func createSecretTmpl(name string) (string, error) {
	tmplData := map[string]string{
		"name": name,
	}
	tmpl, err := template.New("main").Option("missingkey=error").Parse(secretTemplate)
	if err != nil {
		return "", fmt.Errorf("secrets: create secret: %v", err.Error())
	}
	templatedSecret := bytes.Buffer{}
	err = tmpl.Execute(&templatedSecret, tmplData)
	if err != nil {
		return "", fmt.Errorf("secrets: create secret: %v", err.Error())
	}
	filenameCheck := filepath.Join(config.Global.WorkingDir, name+".yaml")
	if _, err := os.Stat(filenameCheck); os.IsNotExist(err) {
		err = ioutil.WriteFile(filenameCheck, templatedSecret.Bytes(), fs.ModePerm)
		if err != nil {
			return "", fmt.Errorf("secrets: create secret: %v", err.Error())
		}
		return filenameCheck, nil
	}
	f, err := ioutil.TempFile(config.Global.WorkingDir, name+"_*.yaml")
	if err != nil {
		return "", fmt.Errorf("secrets: create secret: %v", err.Error())
	}
	defer f.Close()
	_, err = f.Write(templatedSecret.Bytes())
	if err != nil {
		return "", fmt.Errorf("secrets: create secret: %v", err.Error())
	}
	return f.Name(), nil
}

var secretTemplate = `name: {{ .name }}
kind: secret
driver: aws_ssm
spec: 
    region: {{ "{{ .project.variables.region }}" }}
    # secret name in aws ssm 
    ssm_secret: {{ .name }}
`
