package main

import (
	"io/ioutil"

	"github.com/apex/log"
	"gopkg.in/yaml.v2"
)

func main() {
	// Init config (args)

	globalConfigInit()
	loggingInit()

	// Read config ./.test-local/aws.conf.yaml and set global environment variables for aws access.
	// Only for local tests.
	setGlobalBashEnv()

	err := providersLoad()
	if err != nil {
		log.Fatal(err.Error())
	}

	// Get manifests list.
	mList, err := getManifestsList(globalConfig.ClusterConfigsPath)
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
	clusterSpec.ClusterFullName, err = getClusterFullName(clusterSpec.Name, globalConfig.GitRepoName)
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
	} else {
		err = prov.Destroy()
	}
	return err
}
