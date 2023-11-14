package config

import (
	"os"
	"os/user"
	"path/filepath"

	"github.com/apex/log"
	"github.com/shalb/cluster.dev/pkg/colors"
	"github.com/shalb/cluster.dev/pkg/logging"
)

// Version - git tag from compiller
var Version string

// BuildTimestamp - build date from compiller
var BuildTimestamp string

var Interrupted bool

type SubCmd int

const (
	Plan SubCmd = iota
	Apply
	Destroy
	Build
	Clear
)

type CdevUsage struct {
	Operation           string `json:"operation"`
	OperationResult     string `json:"operationResult"`
	AbsoluteTime        string `json:"absoluteTime"`
	RealTime            string `json:"reaLTime"`
	ProjectID           string `json:"projectID"`
	BackendType         string `json:"backendType"`
	CdevVersion         string `json:"cdevVersion"`
	OsVersion           string `json:"osVersion"`
	ProcessedUnitsCount uint   `json:"processedUnitsCount"`
}

// ConfSpec type for global config.
type ConfSpec struct {
	ProjectConfigsPath string
	LogLevel           string
	ProjectConfig      string
	Version            string
	Build              string
	WorkDir            string
	WorkingDir         string
	TraceLog           bool
	MaxParallel        int
	PluginsCacheDir    string
	UseCache           bool
	OptFooTest         bool
	IgnoreState        bool
	NotLoadState       bool
	ShowTerraformPlan  bool
	StateCacheDir      string
	TemplatesCacheDir  string
	CacheDir           string
	NoColor            bool
	Force              bool
	Interactive        bool
	OutputJSON         bool
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
		colors.SetColored(false)
	}
	logging.InitLogLevel(Global.LogLevel, Global.TraceLog)
	Global.ProjectConfigsPath = curPath
	Global.WorkDir = filepath.Join(curPath, ".cluster.dev")
	Global.CacheDir = filepath.Join(Global.WorkDir, "cache/")
	Global.StateCacheDir = filepath.Join(Global.WorkDir, "cache/")
	Global.TemplatesCacheDir = filepath.Join(Global.WorkDir, "templates")
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
	Interrupted = false
}

// getEnv Helper for args parse.
func getEnv(key string, defaultVal string) string {
	if envVal, ok := os.LookupEnv(key); ok {
		return envVal
	}
	return defaultVal
}
