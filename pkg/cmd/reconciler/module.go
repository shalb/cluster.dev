package reconciler

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/apex/log"

	json "github.com/json-iterator/go"
	"github.com/rodaine/hclencoder"
)

type Module struct {
	InfraPtr     *Infrastructure
	ProjectPtr   *Project
	Name         string
	Type         string
	Source       string
	Inputs       map[string]interface{}
	Dependencies []Dependency
}

type Dependency struct {
	Module *Module
	Output string
}

func (m *Module) GenMainCodeBlockHCL() ([]byte, error) {
	type ModuleVars map[string]interface{}

	type HCLModule struct {
		Name       string `hcl:",key"`
		ModuleVars `hcl:",squash"`
	}
	type Config struct {
		Mod HCLModule `hcl:"module"`
	}

	inp, err := json.Marshal(m.Inputs)
	if err != nil {
		log.Fatal(err.Error())
	}
	unmInputs := ModuleVars{}
	err = json.Unmarshal(inp, &unmInputs)
	if err != nil {
		log.Fatal(err.Error())
	}

	unmInputs["source"] = m.Source
	mod := HCLModule{
		Name:       m.Name,
		ModuleVars: unmInputs,
	}

	input := Config{
		Mod: mod,
	}
	return hclencoder.Encode(input)

}

func (m *Module) GenBackendCodeBlockHCL(name string) ([]byte, error) {
	type BackendSpec struct {
		Bucket string `hcl:"bucket"`
		Key    string `hcl:"key"`
		Region string `hcl:"region"`
	}

	type BackendConfig struct {
		BlockKey    string `hcl:",key"`
		BackendSpec `hcl:",squash"`
	}

	type Terraform struct {
		Backend BackendConfig `hcl:"backend"`
		ReqVer  string        `hcl:"required_version"`
	}

	type Config struct {
		TfBlock Terraform `hcl:"terraform"`
	}

	bSpeck := BackendSpec{
		Bucket: name,
		Key:    fmt.Sprintf("%s/%s", m.InfraPtr.Name, m.Name),
		Region: "us-east1",
	}

	tf := Terraform{
		Backend: BackendConfig{
			BlockKey:    "s3",
			BackendSpec: bSpeck,
		},
		ReqVer: "~> 0.13",
	}

	input := Config{
		TfBlock: tf,
	}
	return hclencoder.Encode(input)

}

func (m *Module) GenRemoteStateCodeBlockHCL(name string) ([]byte, error) {

	type BackendSpec struct {
		Bucket string `hcl:"bucket"`
		Key    string `hcl:"key"`
		Region string `hcl:"region"`
	}

	type Data struct {
		KeyRemState  string      `hcl:",key"`
		KeyStateName string      `hcl:",key"`
		Backend      string      `hcl:"backend"`
		Config       BackendSpec `hcl:"config"`
	}

	type Config struct {
		TfBlock []Data `hcl:"data"`
	}

	input := Config{}

	for _, dep := range m.Dependencies {
		tf := Data{
			KeyRemState:  "terraform_remote_state",
			KeyStateName: fmt.Sprintf("%s-%s", dep.Module.InfraPtr.Name, dep.Module.Name),
			Config: BackendSpec{
				Bucket: name,
				Key:    fmt.Sprintf("%s/%s", dep.Module.InfraPtr.Name, dep.Module.Name),
				Region: "us-east1",
			},
			Backend: "s3",
		}

		input.TfBlock = append(input.TfBlock, tf)
	}

	return hclencoder.Encode(input)

}

func findModule(infra, name string, modsList map[string]*Module) *Module {
	mod, exists := modsList[fmt.Sprintf("%s.%s", infra, name)]
	// log.Printf("Check Mod: %s, exists: %v, list %v", name, exists, modsList)
	if !exists {
		return nil
	}
	return mod
}

func (m *Module) checkDependMarker(data reflect.Value) (reflect.Value, error) {
	subVal := reflect.ValueOf(data.Interface())
	resString := subVal.String()
	for key, marker := range m.ProjectPtr.DependencyMarkers {
		if strings.Contains(resString, key) {
			if marker.InfraName == "this" {
				marker.InfraName = m.InfraPtr.Name
			}
			modKey := fmt.Sprintf("%s.%s", marker.InfraName, marker.ModuleName)
			depModule, exists := m.ProjectPtr.Modules[modKey]
			if !exists {
				return reflect.ValueOf(nil), fmt.Errorf("Depend module does not exists. Src: '%s.%s', depend: '%s'", m.InfraPtr.Name, m.Name, modKey)
			}
			m.Dependencies = append(m.Dependencies, Dependency{
				Module: depModule,
				Output: marker.Output,
			})
			remoteStateRef := fmt.Sprintf("${data.terraform_remote_state.%s-%s.%s}", marker.InfraName, marker.ModuleName, marker.Output)
			replacer := strings.NewReplacer(key, remoteStateRef)
			resString = replacer.Replace(resString)
			return reflect.ValueOf(resString), nil
		}
	}
	return subVal, nil
}

func (m *Module) checkYAMLBlockMarker(data reflect.Value) (reflect.Value, error) {
	subVal := reflect.ValueOf(data.Interface())
	for hash := range m.ProjectPtr.InsertYAMLMarkers {
		if subVal.String() == hash {
			return reflect.ValueOf(m.ProjectPtr.InsertYAMLMarkers[hash]), nil
		}
	}
	return subVal, nil
}

// ReplaceMarkers replace all templated markers with values.
func (m *Module) ReplaceMarkers() error {
	err := processingMarkers(m.Inputs, m.checkYAMLBlockMarker)
	if err != nil {
		return err
	}
	err = processingMarkers(m.Inputs, m.checkDependMarker)
	if err != nil {
		return err
	}
	return nil
}
