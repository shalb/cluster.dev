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
func PullKubeConfigOnce(clusterName, doRegion string) ([]byte, error) {
	bash, err := executor.NewBashRunner("")
	if err != nil {
		return nil, err
	}
	kubeConfigName := fmt.Sprintf("kubeconfig_%s", clusterName)
	kubeConfigFilePath := filepath.Join("/tmp", kubeConfigName)
	pullCommand := fmt.Sprintf("s3cmd get %s s3://%s/%s --host=%s.digitaloceanspaces.com --host-bucket=%%(bucket)s.%s.digitaloceanspaces.com", kubeConfigFilePath, clusterName, kubeConfigName, doRegion, doRegion)
	stdout, stderr, err := bash.RunMutely(pullCommand)
	if err != nil {
		log.Debugf("digitalocean s3 cp output:\n%s\nError:\n%s", stdout, stderr)
		return nil, err
	}
	defer os.Remove(kubeConfigFilePath)
	log.Debugf("Kubeconfig pulled. digitalocean s3 cp output: \n%s", stdout)
	kubeConfig, err := ioutil.ReadFile(kubeConfigFilePath)
	if err != nil {
		return nil, err
	}
	return kubeConfig, nil
}

// PullKubeConfig retry pull kube config every 5 sec until timeout.
func PullKubeConfig(clusterName, doRegion string, timeout time.Duration) ([]byte, error) {
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
			kubeConfig, err := PullKubeConfigOnce(clusterName, doRegion)
			if err == nil {
				return kubeConfig, nil
			}
			log.Info("Kubernetes cluster is not ready yet. Will retry after 5 seconds...")
		}
	}
}

// PushKubeConfig upload kube config to s3.
func PushKubeConfig(clusterName, doRegion string, kubeConfig []byte) error {
	bash, err := executor.NewBashRunner("", GetAwsAuthEnv()...)
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
	uploadCommand := fmt.Sprintf("s3cmd put %s s3://%s/%s --host=%s.digitaloceanspaces.com --host-bucket='%%(bucket)s.%s.digitaloceanspaces.com'", kubeConfigFilePath, clusterName, kubeConfigName, doRegion, doRegion)
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

// GetAwsAuthEnv return array with aws auth environments, get there fom do auth env. Needed for s3cmd.
func GetAwsAuthEnv() []string {
	spacesKey := os.Getenv("SPACES_ACCESS_KEY_ID")
	spacesSecret := os.Getenv("SPACES_SECRET_ACCESS_KEY")
	res := []string{
		fmt.Sprintf("AWS_ACCESS_KEY_ID=%s", spacesKey),
		fmt.Sprintf("AWS_SECRET_ACCESS_KEY=%s", spacesSecret),
	}
	return res
}
