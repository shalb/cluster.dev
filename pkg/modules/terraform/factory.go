package terraform

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
	log.Debug("Registering module driver 'terraform'")
	if err := project.RegisterModuleDriverFactory(&modDrv, "terraform"); err != nil {
		log.Trace("Can't register module driver 'terraform'.")
	}
}
