package config

import (
	"flag"
	"os"
	"path/filepath"
	"strings"

	"github.com/apex/log"
)

// ConfSpec type for global config.
type ConfSpec struct {
	GitProvider        string
	GitRepoName        string
	GitRepoRoot        string
	ClusterConfigsPath string
	LogLevel           string
	ProjectRoot        string
	ClusterConfig      string
}

// Global config for executor.
var Global ConfSpec

// set global config values.
func init() {

	// Read flags.
	// Read debug option ( --debug )
	flag.StringVar(&Global.LogLevel, "log-level", getEnv("VERBOSE_LVL", "info"), "Set the logging level (\"debug\"|\"info\"|\"warn\"|\"error\"|\"fatal\") (default \"info\")")
	flag.StringVar(&Global.ClusterConfig, "config", "", "Define cluster config. If empty - reconciler will use all configs by mask ./cluster.dev/*.yaml .")
	curPath, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get current directory: %s", err.Error())
	}
	Global.ClusterConfigsPath = filepath.Join(curPath, ".cluster.dev")
	// Parse args.
	flag.Parse()

	// Detect git provider and set config vars.
	detectGitProvider(&Global)

	// Detect project root dir.
	var ok bool
	if Global.ProjectRoot, ok = os.LookupEnv("PRJ_ROOT"); !ok {
		Global.ProjectRoot, err = os.Getwd()
		if err != nil {
			log.Fatalf("Can't detect project root dir: %s", err.Error())
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

func detectGitProvider(config *ConfSpec) {
	if config.GitRepoName = getEnv("GITHUB_REPOSITORY", ""); config.GitRepoName != "" {
		config.GitProvider = "github"
		config.GitRepoRoot = getEnv("GIT_REPO_ROOT", "")
	} else if config.GitRepoName = getEnv("CI_PROJECT_PATH", ""); config.GitRepoName != "" {
		config.GitProvider = "github"
		config.GitRepoRoot = getEnv("CI_PROJECT_DIR", "")
	} else if config.GitRepoName = getEnv("BITBUCKET_GIT_HTTP_ORIGIN", ""); config.GitRepoName != "" {
		config.GitProvider = "github"
		config.GitRepoRoot = getEnv("BITBUCKET_CLONE_DIR", "")
		config.GitRepoName = strings.ReplaceAll(config.GitRepoName, "http://bitbucket.org/", "")
	} else {
		config.GitProvider = "none"
		config.GitRepoRoot = "./"
		config.GitRepoName = "local/local"
	}
}
