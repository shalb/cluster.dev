package do

import (
	"fmt"

	"github.com/apex/log"
	"github.com/shalb/cluster.dev/pkg/project"
	"gopkg.in/yaml.v2"
)

// Factory factory for do backends.
type Factory struct{}

// New creates the new do backend.
func (f *Factory) New(config []byte, name string) (project.Backend, error) {
	bk := BackendDo{name: name}
	err := yaml.Unmarshal(config, &bk)
	if err != nil {
		return nil, err
	}
	if bk.AccessKey != "" && bk.SecretKey != "" {
		return &bk, nil
	}
	if bk.AccessKey == "" && bk.SecretKey == "" {
		return &bk, nil
	}
	return nil, fmt.Errorf("error in backend '%v' configuration", bk.Name())
}

func init() {
	log.Debug("Registering backend provider do..")
	if err := project.RegisterBackendFactory(&Factory{}, "do"); err != nil {
		log.Fatalf("Can't register backend provider do. %v", err.Error())
	}
}
