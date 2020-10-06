package provisioner

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/apex/log"
	"github.com/shalb/cluster.dev/internal/executor"
)

// PullKubeConfigOnce download kubeconfig from s3 and return it.
func PullKubeConfigOnce(clusterName string) ([]byte, error) {
	bash, err := executor.NewBashRunner("")
	if err != nil {
		return nil, err
	}
	kubeConfigName := fmt.Sprintf("kubeconfig_%s", clusterName)
	kubeConfigFilePath := filepath.Join("/tmp", kubeConfigName)
	pullCommand := fmt.Sprintf("aws s3 cp s3://%s/%s %s", clusterName, kubeConfigName, kubeConfigFilePath)
	stdout, stderr, err := bash.RunMutely(pullCommand)
	if err != nil {
		log.Debugf("aws s3 cp output:\n%s\nError:\n%s", stdout, stderr)
		return nil, err
	}
	defer os.Remove(kubeConfigFilePath)
	log.Debugf("Kubeconfig pulled. aws s3 cp output: \n%s", stdout)
	kubeConfig, err := ioutil.ReadFile(kubeConfigFilePath)
	if err != nil {
		return nil, err
	}
	return kubeConfig, nil
}

// PullKubeConfig retry pull kube config every 5 sec until timeout.
func PullKubeConfig(clusterName string, timeout time.Duration) ([]byte, error) {
	tick := time.NewTicker(5 * time.Second)
	tm := time.NewTimer(timeout)
	defer tick.Stop()
	defer tm.Stop()
	for {
		select {
		case <-tm.C:
			// Timeout
			return nil, fmt.Errorf("kube config pull error: timeout")
		// Wait for tick.
		case <-tick.C:
			kubeConfig, err := PullKubeConfigOnce(clusterName)
			if err == nil {
				return kubeConfig, nil
			}
			log.Info("Minikube cluster is not ready yet. Will retry after 5 seconds...")
		}
	}
}

// PushKubeConfig upload kube config to s3.
func PushKubeConfig(clusterName string, kubeConfig []byte) error {
	bash, err := executor.NewBashRunner("")
	bash.LogLabels = append(bash.LogLabels, fmt.Sprintf("cluster='%s'", clusterName))
	if err != nil {
		return err
	}
	kubeConfigName := fmt.Sprintf("kubeconfig_%s", clusterName)
	kubeConfigFilePath := filepath.Join("/tmp", kubeConfigName)

	// Write kubeconfig to standart path.
	if err = ioutil.WriteFile(kubeConfigFilePath, kubeConfig, os.ModePerm); err != nil {
		return err
	}
	defer os.Remove(kubeConfigFilePath)
	uploadCommand := fmt.Sprintf("aws s3 cp %s s3://%s/%s", kubeConfigFilePath, clusterName, kubeConfigName)
	return bash.Run(uploadCommand)
}

// CheckKubeAccess check connection to kubernetes with kubeConfig.
func CheckKubeAccess(kubeConfig []byte) error {
	// Create tmp dir.

	kub, err := executor.NewKubectlRunner("/tmp", kubeConfig)
	if err != nil {
		return err
	}
	defer kub.Clear()
	return kub.Run("version")
}
