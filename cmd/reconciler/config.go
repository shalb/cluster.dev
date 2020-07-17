package main

import (
	"flag"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/apex/log"
	"github.com/go-yaml/yaml"
	"github.com/romanprog/c-dev/executor"
)

// ConfSpec type for global config.
type ConfSpec struct {
	GitProvider        string
	GitRepoName        string
	GitRepoRoot        string
	ClusterConfigsPath string
	LogLevel           string
}

// Configuration args.
var globalConfig ConfSpec

func globalConfigInit() {

	var err error
	// Read flags.
	// Read debug option ( --debug )
	flag.StringVar(&globalConfig.LogLevel, "log-level", getEnv("VERBOSE_LVL", "info"), "Set the logging level (\"debug\"|\"info\"|\"warn\"|\"error\"|\"fatal\") (default \"info\")")

	curPath, err := os.Getwd()
	globalConfig.ClusterConfigsPath, err = os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get current directory: %s", err.Error())
	}
	globalConfig.ClusterConfigsPath = filepath.Join(curPath, ".cluster.dev")
	// Parse args.
	flag.Parse()

	// Detect git provider and set config vars.
	detectGitProvider(&globalConfig)

	log.Debugf("Config: %+v\n", globalConfig)
}

// getEnv Helper for args parse.
func getEnv(key string, defaultVal string) string {
	if envVal, ok := os.LookupEnv(key); ok {
		return envVal
	}
	return defaultVal
}

// TODO remove this
func setGlobalBashEnv() error {
	envFile := ".test-local/aws.conf.yaml"
	var envs struct {
		Environment []string `yaml:"environments"`
	}
	yamlData, err := ioutil.ReadFile(envFile)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(yamlData, &envs)
	if err != nil {
		return err
	}
	log.Debugf("%v", envs.Environment)
	executor.Env = envs.Environment
	return nil
}
