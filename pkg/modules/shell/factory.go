package shell

import (
	"github.com/apex/log"
	"github.com/shalb/cluster.dev/pkg/project"
)

// Factory factory for s3 backends.
type Factory struct {
}

// New creates new module driver factory.
func (f *Factory) New(prj *project.Project) project.ModuleDriver {
	return &TFModuleDriver{
		projectPtr: prj,
	}
}

func init() {
	modDrv := Factory{}
	log.Debugf("Registering module driver '%s'", moduleTypeKey)
	if err := project.RegisterModuleDriverFactory(&modDrv, moduleTypeKey); err != nil {
		log.Fatalf("Can't register module driver '%s'", moduleTypeKey)
	}
}
