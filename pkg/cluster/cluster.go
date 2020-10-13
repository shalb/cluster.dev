package cluster

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/apex/log"
	"github.com/aybabtme/rgbterm"
	"github.com/shalb/cluster.dev/internal/config"
	"github.com/shalb/cluster.dev/internal/utils"
	"github.com/shalb/cluster.dev/pkg/apps"
	"gopkg.in/yaml.v2"
)

// Config - cluster data from cluster manifest.
type Config struct {
	Name            string          `yaml:"name"`
	Installed       bool            `yaml:"installed"`
	Addons          map[string]bool `yaml:"addons,omitempty"`
	Apps            []string        `yaml:"apps,omitempty"`
	ProviderConfig  interface{}     `yaml:"provider"`
	ClusterFullName string
}

// Cluster impl.
type Cluster struct {
	Config   Config
	state    *State
	Provider Provider
}

// State - cluster state which is changed by module and provisioners.
type State struct {
	KubeConfig       []byte
	KubeAccessInfo   string
	AddonsAccessInfo string
}

// New cluster.
func New(clusterConfig []byte) (*Cluster, error) {

	clusterCfg, err := parseClusterManifest(clusterConfig)
	if err != nil {
		return nil, err
	}
	log.Debugf("Cluster config: %+v", clusterCfg)
	st := &State{}
	provider, err := NewProvider(clusterCfg, st)
	cluster := &Cluster{
		Config:   clusterCfg,
		Provider: provider,
		state:    st,
	}
	if err != nil {
		return nil, err
	}
	return cluster, nil
}

// Reconcile cluster.
func (c *Cluster) Reconcile() error {

	var err error
	if c.Config.Installed {
		log.Infof("Deploying cluster '%s'", c.Config.ClusterFullName)
		err = c.Provider.Deploy()
		if err == nil {
			// Output access info in green color.
			log.Info(rgbterm.FgString(c.state.KubeAccessInfo, 51, 255, 0))
			log.Info(rgbterm.FgString(c.state.AddonsAccessInfo, 51, 255, 0))
		}
	} else {
		log.Infof("Destroying cluster '%s'", c.Config.ClusterFullName)
		err = c.Provider.Destroy()
	}
	if err != nil {
		return err
	}
	if c.Config.Installed {
		c.deployApps()
	}
	return nil
}

func (c *Cluster) deployApps() {
	for _, app := range c.Config.Apps {
		curPath, err := os.Getwd()
		if err != nil {
			log.Fatalf("Failed to get current directory: %s", err.Error())
		}
		appFullPath := filepath.Join(curPath, app)
		log.Infof("Deploying app: %s", appFullPath)
		err = apps.Deploy(appFullPath, c.state.KubeConfig)
		if err != nil {
			log.Errorf("App deploying error: %s", err.Error())
		}
	}
}

// parseClusterManifest parse cluster config.
func parseClusterManifest(clusterConfig []byte) (Config, error) {
	var clusterSpec Config
	err := yaml.Unmarshal(clusterConfig, &clusterSpec)
	if err != nil {
		return clusterSpec, err
	}
	clusterSpec.ClusterFullName, err = getClusterFullName(clusterSpec.Name, config.Global.GitRepoName)
	return clusterSpec, nil
}

func getClusterFullName(clusterName, gitRepoName string) (string, error) {
	splitted := strings.Split(gitRepoName, "/")
	var strTmp string
	if len(splitted) > 1 {
		strTmp = splitted[0]
	}
	if strTmp == "" {
		return "", fmt.Errorf("can't set cluster fullname from git repo name. GitRepoName is '%s'", gitRepoName)
	}
	// Prepate string.
	strTmp = strings.ToLower(fmt.Sprintf("%s-%s", clusterName, strTmp))
	// Truncate string.
	strTmp = utils.TruncateString(strTmp, 63)
	return strTmp, nil
}
