package sops

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/apex/log"
	"github.com/shalb/cluster.dev/internal/config"
	"github.com/shalb/cluster.dev/internal/project"
	"github.com/shalb/cluster.dev/pkg/executor"
	"github.com/shalb/cluster.dev/pkg/sopstools"
	"github.com/shalb/cluster.dev/pkg/utils"
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
	secretSpec, err := utils.ReadYAML(decryptedRaw)
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
	runner, err := executor.NewExecutor(config.Global.WorkingDir, &config.Interrupted)
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

func (s *sopsDriver) Create(files map[string][]byte) error {
	runner, err := executor.NewExecutor(config.Global.WorkingDir, &config.Interrupted)
	if err != nil {
		return err
	}
	if len(files) != 1 {
		return fmt.Errorf("create sops secret: expected 1 file, received %v", len(files))
	}
	for fn, data := range files {
		filename, err := saveTmplToFile(fn, data)
		if err != nil {
			return fmt.Errorf("create sops secret: %v", err.Error())
		}
		command := fmt.Sprintf("sops -e --encrypted-regex ^encrypted_data$ -i %s", filename)
		err = runner.RunWithTty(command)
		if err != nil {
			os.Remove(filename)
			return fmt.Errorf("create sops secret: %v", err.Error())
		}
		command = fmt.Sprintf("sops %s", filename)
		err = runner.RunWithTty(command)
		if err != nil && err.Error() != "exit status 200" {
			os.Remove(filename)
			log.Debugf("err %+v", err)
			return fmt.Errorf("create sops secret: %v", err.Error())
		}
	}
	return nil
}

func saveTmplToFile(name string, data []byte) (string, error) {
	filenameCheck := filepath.Join(config.Global.WorkingDir, name)
	if _, err := os.Stat(filenameCheck); os.IsNotExist(err) {
		err = os.WriteFile(filenameCheck, data, fs.ModePerm)
		if err != nil {
			return "", err
		}
		return filenameCheck, nil
	}
	f, err := os.CreateTemp(config.Global.WorkingDir, "*_"+name)
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
