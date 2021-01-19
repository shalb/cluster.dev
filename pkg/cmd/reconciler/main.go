package reconciler

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/apex/log"

	"github.com/shalb/cluster.dev/internal/config"
	"github.com/shalb/cluster.dev/pkg/project"
)

// Run main process.
func Run() {

	if config.Global.OnlyPrintVersion {
		fmt.Printf("Version: %s\nBuild: %s\n", config.Global.Version, config.Global.Build)
		return
	}

	projConf, manifests := getManifests(config.Global.ClusterConfigsPath)
	project, err := project.NewProject(projConf, manifests)
	if err != nil {
		log.Fatal(err.Error())
	}
	err = project.GenCode()
	if err != nil {
		log.Fatal(err.Error())
	}
	if config.Global.SubCommand == config.Build {
		return
	}
	if config.Global.SubCommand == config.Plan {
		err = project.Plan()
		if err != nil {
			log.Fatal(err.Error())
		}
		return
	}
	if config.Global.SubCommand == config.Apply {
		err = project.Apply()
		if err != nil {
			log.Fatal(err.Error())
		}
		return
	}
	if config.Global.SubCommand == config.Destroy {
		err = project.Destroy()
		if err != nil {
			log.Fatal(err.Error())
		}
		return
	}
	return
}

// Return project conf and slice of others config files.
func getManifests(path string) ([]byte, [][]byte) {

	var files []string
	var projectConf []byte
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
		manifest := []byte{}
		if filepath.Base(file) == project.ConfigFileName {
			projectConf, err = ioutil.ReadFile(file)
		} else {
			manifest, err = ioutil.ReadFile(file)
			manifests[i] = manifest
		}
		if err != nil {
			log.Fatalf("error while reading %v: %v", file, err)
		}
	}
	return projectConf, manifests
}
