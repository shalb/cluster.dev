package reconciler

import (
	"io/ioutil"
	"path/filepath"

	"github.com/apex/log"
	"github.com/shalb/cluster.dev/internal/config"
	"github.com/shalb/cluster.dev/pkg/cluster"
)

// Run main process.
func Run() {

	manifests := getManifests(config.Global.ClusterConfigsPath)
	for _, cManifest := range manifests {
		cluster, err := cluster.New(cManifest)
		if err != nil {
			log.Fatalf("Config file:\n %s \nError: %s\n", string(cManifest), err.Error())
		}
		// log.Printf("Cluster `%v`: starting reconciliation\n", cluster.GetConfig().Name)
		if err = cluster.Reconcile(); err != nil {
			log.Fatalf("Error occurred during reconciliation of cluster %s", err.Error())
		}
	}

	return
}

func getManifests(path string) [][]byte {

	var files []string
	var err error
	if config.Global.ClusterConfig != "" {
		files = append(files, config.Global.ClusterConfig)
	} else {
		files, err = filepath.Glob(path + "/*.yaml")
		if err != nil {
			log.Fatalf("cannot read directory %v: %v", path, err)
		}
	}
	if len(files) == 0 {
		log.Fatalf("no manifest found in %v", path)
	}

	manifests := make([][]byte, len(files))
	for i, file := range files {
		manifest, err := ioutil.ReadFile(file)
		if err != nil {
			log.Fatalf("error while reading %v: %v", file, err)
		}
		manifests[i] = manifest
	}
	return manifests
}
