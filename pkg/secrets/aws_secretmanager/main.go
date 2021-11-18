package aws

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/apex/log"
	"github.com/shalb/cluster.dev/pkg/aws"
	"github.com/shalb/cluster.dev/pkg/config"
	"github.com/shalb/cluster.dev/pkg/executor"
	"github.com/shalb/cluster.dev/pkg/project"
	"github.com/shalb/cluster.dev/pkg/utils"
	"gopkg.in/yaml.v3"
)

const secretmanagerKey = "aws_secretmanager"

type smDriver struct{}

type secretmanagerSpec struct {
	Region     string      `yaml:"region"`
	SecretName string      `yaml:"aws_secret_name"`
	Data       interface{} `yaml:"secret_data,omitempty"`
}

func (s *smDriver) Read(rawData []byte) (name string, data interface{}, err error) {
	secretSpec, err := utils.ReadYAML(rawData)
	if err != nil {
		return
	}
	name, ok := secretSpec["name"].(string)
	if !ok {
		err = fmt.Errorf("aws_secretmanager: secret must contain string field 'name'")
		return
	}
	sp, ok := secretSpec["spec"].(map[string]interface{})
	if !ok {
		err = fmt.Errorf("aws_secretmanager: secret must contain field 'region'")
		return
	}
	specRaw, err := yaml.Marshal(sp)
	if err != nil {
		err = fmt.Errorf("aws_secretmanager: can't parse secret '%v' spec %v", name, err)
		return
	}
	var spec secretmanagerSpec
	err = yaml.Unmarshal(specRaw, &spec)
	if err != nil {
		err = fmt.Errorf("aws_secretmanager: can't parse secret '%v' spec %v", name, utils.ResolveYamlError(specRaw, err))
		return
	}
	if spec.Region == "" || spec.SecretName == "" {
		if err != nil {
			err = fmt.Errorf("aws_secretmanager: can't parse secret '%v', fields 'spec.region' and 'spec.secret_name' are required", name)
			return
		}
	}
	data, err = aws.GetSecret(spec.Region, spec.SecretName)
	if err != nil {
		return "", nil, err
	}

	log.Debugf("%+v", data)
	return
}

func (s *smDriver) Key() string {
	return secretmanagerKey
}

func init() {
	err := project.RegisterSecretDriver(&smDriver{}, secretmanagerKey)
	if err != nil {
		log.Fatalf("secrets: secretmanager driver init: %v", err.Error())
	}
}

func (s *smDriver) Edit(sec project.Secret) error {
	runner, err := executor.NewExecutor(config.Global.WorkingDir)
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

func (s *smDriver) Create(files map[string][]byte) error {
	runner, err := executor.NewExecutor(config.Global.WorkingDir)
	if err != nil {
		return fmt.Errorf("create  secret: %v", err.Error())
	}
	if len(files) != 1 {
		return fmt.Errorf("create sops secret: expected 1 file, received %v", len(files))
	}
	for fn, data := range files {
		filename, err := saveTmplToFile(fn, data)
		if err != nil {
			return fmt.Errorf("create sops secret: %v", err.Error())
		}
		editor := os.Getenv("EDITOR")
		if editor == "" {
			editor = "vi"
		}
		command := fmt.Sprintf("%s %s", editor, filename)
		err = runner.RunWithTty(command)
		if err != nil {
			os.RemoveAll(filename)
			return fmt.Errorf("secrets: create secret: %v", err.Error())
		}
	}
	return nil
}

func saveTmplToFile(name string, data []byte) (string, error) {
	filenameCheck := filepath.Join(config.Global.WorkingDir, name)
	if _, err := os.Stat(filenameCheck); os.IsNotExist(err) {
		err = ioutil.WriteFile(filenameCheck, data, fs.ModePerm)
		if err != nil {
			return "", err
		}
		return filenameCheck, nil
	}
	f, err := ioutil.TempFile(config.Global.WorkingDir, "*_"+name)
	if err != nil {
		return "", err
	}
	defer f.Close()
	_, err = f.Write(data)
	if err != nil {
		return "", err
	}
	return f.Name(), nil
}
