package executor

import (
	"fmt"
)

// KubectlRunner - exec kubectl commands.
type KubectlRunner struct {
	kubeConfigPath string
	workingDir     string
	bashRunner     *BashRunner
}

// NewKubectlRunner create kubectl runner.
func NewKubectlRunner(workingDir, kConfPath string) (*KubectlRunner, error) {
	var k KubectlRunner
	var err error
	k.bashRunner, err = NewBashRunner(workingDir)
	if err != nil {
		return nil, err
	}
	k.workingDir = workingDir
	k.kubeConfigPath = kConfPath
	return &k, nil
}

// Run - common function to prepare and run kubectl commands in bash shell.
func (k *KubectlRunner) Run(args ...string) error {
	// Run command and return result.
	var commonCommand string
	if k.kubeConfigPath != "" {
		commonCommand = fmt.Sprintf("kubectl --kubeconfig %s", k.kubeConfigPath)
	} else {
		commonCommand = "kubectl"
	}
	for _, arg := range args {
		commonCommand = fmt.Sprintf("%s %s", commonCommand, arg)
	}
	return k.bashRunner.Run(commonCommand)
}

// Version - exec kubectl version.
func (k *KubectlRunner) Version() (string, error) {
	var err error
	o, oerr, err := k.bashRunner.RunMutely("kubectl version")
	if err != nil {
		return "", fmt.Errorf("%s, error output: %s", err.Error(), oerr)
	}
	return o, nil
}
