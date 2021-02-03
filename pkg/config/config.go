package config

import (
	"flag"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/apex/log"
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
	GitProvider        string
	GitRepoName        string
	GitRepoRoot        string
	ClusterConfigsPath string
	LogLevel           string
	ProjectRoot        string
	ClusterConfig      string
	Version            string
	Build              string
	TmpDir             string
	WorkingDir         string
	TraceLog           bool
	OnlyPrintVersion   bool
	MaxParallel        int
	SubCommand         SubCmd
	PluginsCacheDir    string
	UseCache           bool
}

// Global config for executor.
var Global ConfSpec

// set global config values.
func init() {

	// Read flags.
	// Read debug option ( --debug )
	flag.StringVar(&Global.LogLevel, "log-level", getEnv("VERBOSE_LVL", "info"), "Set the logging level (\"debug\"|\"info\"|\"warn\"|\"error\"|\"fatal\") (default \"info\")")
	flag.BoolVar(&Global.OnlyPrintVersion, "version", false, "Print binary version tag.")
	flag.BoolVar(&Global.TraceLog, "trace", false, "Print function trace info in logs.")
	flag.IntVar(&Global.MaxParallel, "max-paraless", 3, "Max parallel module applying")
	var build, apply, plan, destroy bool
	flag.BoolVar(&build, "build", false, "Build project code")
	flag.BoolVar(&apply, "apply", false, "Apply project")
	flag.BoolVar(&plan, "plan", false, "Show terraform plan")
	flag.BoolVar(&destroy, "destroy", false, "Destroy project")
	flag.BoolVar(&Global.UseCache, "cache", false, "Use previously cached build directory")

	curPath, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get current directory: %s", err.Error())
	}
	Global.WorkingDir = curPath
	Global.Version = Version
	Global.Build = BuildTimestamp

	Global.ClusterConfigsPath = curPath
	Global.TmpDir = filepath.Join(curPath, ".cluster.dev")
	// Parse args.
	flag.Parse()

	if Global.OnlyPrintVersion {
		fmt.Printf("Version: %s\nBuild: %s\n", Global.Version, Global.Build)
		os.Exit(0)
	}

	cmdsCount := 0
	if build {
		cmdsCount++
		Global.SubCommand = Build
	}
	if apply {
		cmdsCount++
		Global.SubCommand = Apply
	}
	if plan {
		cmdsCount++
		Global.SubCommand = Plan
	}
	if destroy {
		cmdsCount++
		Global.SubCommand = Destroy
	}
	if cmdsCount > 1 {
		log.Fatal("You should use only one of commands: (-apply|-plan|-destroy|-build)")
	}
	if cmdsCount < 1 {
		log.Fatal("Command require: (-apply|-plan|-destroy|-build)")
	}
	// Detect git provider and set config vars.
	detectGitProvider(&Global)
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
