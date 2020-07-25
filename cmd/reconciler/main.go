package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/apex/log"
	"github.com/shalb/cluster.dev/internal/apps"
	"github.com/shalb/cluster.dev/internal/config"
	"gopkg.in/yaml.v2"
)

func main() {
	// Init config (args)

	config.InitGlobal()

	loggingInit()

	err := providersLoad()
	if err != nil {
		log.Fatal(err.Error())
	}

	// Get manifests list.
	mList, err := getManifestsList(config.Global.ClusterConfigsPath)
	if err != nil {
		log.Fatal(err.Error())
	}

	log.Debugf("Manifests list:\n%+v", mList)
	// Iterate trough provided manifests and reconcile clusters.
	for _, clusterManifest := range mList {
		err = reconcileCluster(clusterManifest)
		if err != nil {
			log.Fatal(err.Error())
		}
	}

}

func reconcileCluster(clusterManifest string) error {
	// Read main yaml file.
	mainYaml, err := ioutil.ReadFile(clusterManifest)
	log.Debug("Read cluster manifest.")
	if err != nil {
		return err
	}
	// Parse yaml file.
	// var parsedYaml interface{}
	log.Debug("Parse cluster manifest.")
	var clusterSpec ClusterSpec
	err = yaml.Unmarshal([]byte(mainYaml), &clusterSpec)
	if err != nil {
		return err
	}

	// Generate a unique name for particular cluster domains, state buckets, etc..
	clusterSpec.ClusterFullName, err = getClusterFullName(clusterSpec.Name, config.Global.GitRepoName)
	if err != nil {
		return err
	}

	// Convert provider subcategory to raw yaml.
	log.Debug("Marshal provider data to string.")
	providerData, err := yaml.Marshal(clusterSpec.ProviderConfig)
	if err != nil {
		return err
	}

	log.Debug("Create provider.")
	// Create new provider.
	prov, err := newProvider(providerData, clusterSpec.ClusterFullName)
	if err != nil {
		return err
	}
	// Deploy or destroy cluster.
	if clusterSpec.Installed {
		err = prov.Deploy()
		kubeConf := prov.GetKubeConfig()
		if kubeConf == "" {
			return fmt.Errorf("provider return empty kube config, can't deploy applications")
		}
		// Save kube config to file.
		kubeConfigPath := filepath.Join("/tmp/", "kubeconfig_"+clusterSpec.ClusterFullName)
		err = ioutil.WriteFile(kubeConfigPath, []byte(kubeConf), os.ModePerm)
		if err != nil {
			return err
		}
		for _, appSubDir := range clusterSpec.Apps {
			appDir := filepath.Join(config.Global.ProjectRoot, appSubDir)
			log.Debugf("Deploying app '%s'...", appDir)
			app, err := apps.New(appDir, kubeConfigPath)
			if err != nil {
				return err
			}
			err = app.Deploy()
			if err != nil {
				return nil
			}
		}
	} else {
		err = prov.Destroy()
	}
	return err
}
