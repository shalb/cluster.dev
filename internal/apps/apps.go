package apps

import (
	"github.com/shalb/cluster.dev/internal/executor"
)

// Application - type for kubernetes apps.
type Application struct {
	dir     string
	kubectl *executor.KubectlRunner
}

// New - create new application.
func New(appPath string, kubeConfigPath string) (*Application, error) {
	var app Application
	kub, err := executor.NewKubectlRunner(appPath, kubeConfigPath)
	if err != nil {
		return nil, err
	}
	app.kubectl = kub
	app.dir = appPath
	return &app, nil

}

// Deploy application (recursive dir apply)
func (a *Application) Deploy() error {
	return a.kubectl.Run("apply", "-f", "./", "--recursive")
}
