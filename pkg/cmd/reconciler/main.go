package reconciler

import (
	"github.com/shalb/cluster.dev/pkg/cluster"
	"io/ioutil"
	"path/filepath"

	"log"

	_ "github.com/shalb/cluster.dev/pkg/provider/aws"
	_ "github.com/shalb/cluster.dev/pkg/provider/aws/provisioner/eks"
	_ "github.com/shalb/cluster.dev/pkg/provider/aws/provisioner/minikube"
)

func Run() {
	manifests := getManifests("./test")
	rs := make([]cluster.Reconciler, len(manifests))
	for i, manifest := range manifests {
		r, err := cluster.NewFrom(manifest)
		if err != nil {
			log.Fatalf("Config file:\n %v \nError: %v\n", string(manifest), err)
		}
		rs[i] = r
	}

	for _, r := range rs {
		log.Printf("Cluster `%v`: starting reconciliation\n", r.GetConfig().Name)
		err := r.Reconcile()
		if err != nil {
			log.Fatalf("Error occured during reconciliation of cluster '%v' : %v", r.GetConfig().Name, err)
		}
		log.Printf("Cluster `%v`: reconciliation finished\n", r.GetConfig().Name)
	}
}

func getManifests(path string) [][]byte {
	files, err := filepath.Glob(path + "/*.yaml")
	if err != nil {
		log.Fatalf("cannot read directory %v: %v", path, err)
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
