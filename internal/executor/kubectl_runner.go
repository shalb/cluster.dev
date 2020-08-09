package executor

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

// KubectlRunner - exec kubectl commands.
type KubectlRunner struct {
	kubeConfigFile string
	tmpDir         string
	workingDir     string
	bashRunner     *BashRunner
}

// NewKubectlRunner create kubectl runner.
func NewKubectlRunner(workingDir string, kubeConfig []byte) (*KubectlRunner, error) {
	var k KubectlRunner
	var err error
	k.bashRunner, err = NewBashRunner(workingDir)
	if err != nil {
		return nil, err
	}
	dir, err := ioutil.TempDir("", "cluster-dev-kube-*")
	if err != nil {
		return nil, err
	}
	// Save cube config to tmp file.
	k.kubeConfigFile = filepath.Join(dir, "kube_config")
	if err := ioutil.WriteFile(k.kubeConfigFile, kubeConfig, os.ModePerm); err != nil {
		return nil, err
	}
	k.workingDir = workingDir
	k.tmpDir = dir
	return &k, nil
}

// Run - common function to prepare and run kubectl commands in bash shell.
func (k *KubectlRunner) Run(args ...string) error {
	// Run command and return result.
	var commonCommand string
	if k.kubeConfigFile != "" {
		commonCommand = fmt.Sprintf("kubectl --kubeconfig %s", k.kubeConfigFile)
	} else {
		commonCommand = "kubectl"
	}
	for _, arg := range args {
		commonCommand = fmt.Sprintf("%s %s", commonCommand, arg)
	}
	return k.bashRunner.Run(commonCommand)
}

// Clear remove tmp dir.
func (k *KubectlRunner) Clear() {
	os.RemoveAll(k.tmpDir)
}

// Write kube config to file and return full path.
func saveConfig(kubeConfig []byte) (string, error) {
	dir, err := ioutil.TempDir("", "cluster-dev-kube-*")
	if err != nil {
		return "", err
	}
	// Save cube config to tmp file.
	kubeConfigFilePath := filepath.Join(dir, "kube_config")
	if err := ioutil.WriteFile(kubeConfigFilePath, kubeConfig, os.ModePerm); err != nil {
		return "", err
	}
	return kubeConfigFilePath, nil
}
