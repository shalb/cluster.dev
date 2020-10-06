package apps

import (
	"github.com/shalb/cluster.dev/internal/executor"
)

// Deploy application (recursive dir apply)
func Deploy(appDir string, kubeConfig []byte) error {
	kub, err := executor.NewKubectlRunner(appDir, kubeConfig)
	if err != nil {
		return err
	}
	defer kub.Clear()
	return kub.Run("apply", "-f", "./", "--recursive")
}
