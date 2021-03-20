package config

import (
	"bytes"
	"fmt"
	"os"
	"os/user"
	"path/filepath"

	"github.com/apex/log"
	"github.com/shalb/cluster.dev/pkg/colors"
	"github.com/shalb/cluster.dev/pkg/logging"
	"gopkg.in/yaml.v3"
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
	IgnoreState        bool
	ShowTerraformPlan  bool
	StateFileName      string
	StateCacheDir      string
	CacheDir           string
	NoColor            bool
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
	if Global.NoColor {
		// Turn off colored output.
		colors.InitColors(false)
	}
	logging.InitLogLevel(Global.LogLevel, Global.TraceLog)
	Global.ClusterConfigsPath = curPath
	Global.TmpDir = filepath.Join(curPath, ".cluster.dev")
	Global.CacheDir = filepath.Join(Global.TmpDir, "cache")
	Global.StateCacheDir = filepath.Join(Global.TmpDir, "state")
	Global.StateFileName = filepath.Join(curPath, "cdev.state")
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

}

// getEnv Helper for args parse.
func getEnv(key string, defaultVal string) string {
	if envVal, ok := os.LookupEnv(key); ok {
		return envVal
	}
	return defaultVal
}

func ReadYAMLObjects(objData []byte) ([]map[string]interface{}, error) {
	objects := []map[string]interface{}{}
	dec := yaml.NewDecoder(bytes.NewReader(objData))
	for {
		var parsedConf = make(map[string]interface{})
		err := dec.Decode(&parsedConf)
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			log.Debugf("can't decode config to yaml: %s", err.Error())
			return nil, fmt.Errorf("can't decode config to yaml: %s", err.Error())
		}
		objects = append(objects, parsedConf)
	}
	return objects, nil
}
