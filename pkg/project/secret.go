package project

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/apex/log"
	"github.com/olekukonko/tablewriter"
	"github.com/shalb/cluster.dev/pkg/config"
)

type secretDriver int

const secretObjKindKey = "secret"

type Secret struct {
	Filename  string
	DriverKey string
	Data      interface{}
}

func (p *Project) readSecrets() error {
	for filename, data := range p.objectsFiles {
		templatedRaw, hasWarn, tmplErr := p.TemplateTry(data)
		if tmplErr != nil && !hasWarn {
			log.Debug(tmplErr.Error())
			return fmt.Errorf("searching for secrets in the project dir: %v", tmplErr.Error())
		}
		secretDriver, err := getRwaSecretInfo(templatedRaw, p)
		if err != nil {
			return fmt.Errorf("searching for secrets in the project dir: %v", err.Error())
		}
		if secretDriver == nil {
			continue
		}
		if hasWarn {
			return fmt.Errorf("searching for secrets in the project dir: %v", tmplErr.Error())
		}
		name, data, err := secretDriver.Read(templatedRaw)
		if err != nil {
			return fmt.Errorf("searching for secrets in the project dir: %v", err.Error())
		}
		if _, exists := p.secrets[name]; exists {
			return fmt.Errorf("searching for secrets in the project dir: duplicated secret name '%v'", name)
		}
		p.secrets[name] = Secret{Filename: filename, DriverKey: secretDriver.Key(), Data: data}
		if _, exists := p.configData["secret"]; !exists {
			p.configData["secret"] = map[string]interface{}{}
		}
		p.configData["secret"].(map[string]interface{})[name] = data
	}

	return nil
}

func (p *Project) fileIsSecret(fn string) bool {
	for _, sec := range p.secrets {
		if sec.Filename == fn {
			return true
		}
	}
	return false
}

func getRwaSecretInfo(data []byte, p *Project) (res SecretDriver, err error) {
	objects, err := ReadYAMLObjects(data)
	if err != nil {
		return
	}
	if len(objects) != 1 {
		return nil, nil
	}
	res, err = getSecretInfo(objects[0])
	if err != nil {
		return
	}

	return
}

func getSecretInfo(obj map[string]interface{}) (res SecretDriver, err error) {
	name, ok := obj["name"].(string)
	if !ok {
		err = fmt.Errorf("secret spec: should contain 'name' field")
		return
	}
	err = checkSecretName(name)
	if err != nil {
		err = fmt.Errorf("secret spec: %v", err.Error())
		return
	}
	kind, ok := obj["kind"].(string)
	if !ok {
		return
	}
	if kind == "secret" {
		driver, ok := obj["driver"].(string)
		if !ok {
			err = fmt.Errorf("secret spec: should contain 'driver' field")
			return
		}
		res, ok = SecretDriversMap[driver]
		if !ok {
			err = fmt.Errorf("secret spec: unknown driver type '%v'", driver)
			return
		}
	}
	return
}

func (p *Project) PrintSecretsList() {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Name", "Driver", "File"})
	for name, secret := range p.secrets {
		relPath, _ := filepath.Rel(config.Global.WorkingDir, secret.Filename)
		table.Append([]string{name, secret.DriverKey, "./" + relPath})
	}
	table.Render()
}

func (p *Project) Create(drvName, name string) error {
	err := checkSecretName(name)
	if err != nil {
		return fmt.Errorf("creating secret: %v", err.Error())
	}
	if _, exists := p.secrets[name]; exists {
		return fmt.Errorf("secret with name '%v' is already exists", name)
	}
	drv, ok := SecretDriversMap[drvName]
	if !ok {
		return fmt.Errorf("unknown secret driver '%v'", drvName)
	}
	return drv.Create(name)
}

func (p *Project) Edit(name string) error {
	if _, exists := p.secrets[name]; !exists {
		return fmt.Errorf("secret '%v' not found", name)
	}
	return SecretDriversMap[p.secrets[name].DriverKey].Edit(p.secrets[name])
}

type SecretDriver interface {
	Read([]byte) (string, interface{}, error)
	Key() string
	Edit(Secret) error
	Create(string) error
}

var SecretDriversMap = map[string]SecretDriver{}

func RegisterSecretDriver(drv SecretDriver, key string) error {
	if _, exists := SecretDriversMap[key]; exists {
		return fmt.Errorf("secret driver is already exists '%v'", key)
	}
	SecretDriversMap[key] = drv
	return nil
}

func checkSecretName(name string) error {
	res := regexp.MustCompile(`^[a-zA-Z][a-zA-Z_0-9]{0,32}$`).MatchString(name)
	if !res {
		return fmt.Errorf("incorrect name '%v', should contain only alphabetic characters a-z, A-Z, 0-9, _ and start with an alphabetic character. Regex to verify: [a-zA-Z][a-zA-Z_0-9]{0,32}", name)
	}
	return nil
}
