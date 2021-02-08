package config

import (
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"

	"github.com/apex/log"
	"github.com/shalb/cluster.dev/pkg/logging"
)

// Version - git tag from compiller
var Version string

// BuildTimestamp - build date from compiller
var BuildTimestamp string

type SubCmd int

const (
	Plan SubCmd = iota
	Apply
	Destroy
	Build
	Clear
)

// ConfSpec type for global config.
type ConfSpec struct {
	ClusterConfigsPath string
	LogLevel           string
	ClusterConfig      string
	Version            string
	Build              string
	TmpDir             string
	WorkingDir         string
	TraceLog           bool
	MaxParallel        int
	PluginsCacheDir    string
	UseCache           bool
	OptFooTest         bool
	Manifests          [][]byte
	ProjectConf        []byte
}

// Global config for executor.
var Global ConfSpec

// InitConfig set global config values.
func InitConfig() {
	curPath, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get current directory: %s", err.Error())
	}
	Global.WorkingDir = curPath
	Global.Version = Version
	Global.Build = BuildTimestamp
	logging.InitLogLevel(Global.LogLevel, Global.TraceLog)
	Global.ClusterConfigsPath = curPath
	Global.TmpDir = filepath.Join(curPath, ".cluster.dev")

	usr, err := user.Current()
	if err != nil {
		log.Fatal(err.Error())
	}
	if err != nil {
		log.Fatal(err.Error())
	}
	Global.PluginsCacheDir = filepath.Join(usr.HomeDir, ".terraform.d/plugin-cache")
	if _, err := os.Stat(Global.PluginsCacheDir); os.IsNotExist(err) {
		err := os.MkdirAll(Global.PluginsCacheDir, 0755)
		if err != nil {
			log.Fatal(err.Error())
		}
	}
	Global.ProjectConf, Global.Manifests = getManifests(Global.ClusterConfigsPath)
}

// getEnv Helper for args parse.
func getEnv(key string, defaultVal string) string {
	if envVal, ok := os.LookupEnv(key); ok {
		return envVal
	}
	return defaultVal
}

// Return project conf and slice of others config files.
func getManifests(path string) ([]byte, [][]byte) {

	var files []string
	var projectConf []byte
	var err error
	if Global.ClusterConfig != "" {
		files = append(files, Global.ClusterConfig)
	} else {
		files, err = filepath.Glob(path + "/*.yaml")
		if err != nil {
			log.Fatalf("cannot read directory %v: %v", path, err)
		}
	}
	if len(files) < 2 {
		log.Fatalf("no manifest found in %v", path)
	}

	manifests := make([][]byte, len(files)-1)
	for i, file := range files {
		manifest := []byte{}
		if filepath.Base(file) == "project.yaml" {
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
