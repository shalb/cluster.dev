package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/apex/log"
)

// ClusterSpec - cluster data from cluster manifest.
type ClusterSpec struct {
	Name            string          `yaml:"name"`
	Installed       bool            `yaml:"installed"`
	Addons          map[string]bool `yaml:"addons,omitempty"`
	Apps            []string        `yaml:"apps,omitempty"`
	ProviderConfig  interface{}     `yaml:"provider"`
	ClusterFullName string
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

func truncateString(src string, maxLen int) string {

	if len(src) <= maxLen {
		return src
	}

	res := src[0 : maxLen-1]
	return res
}

func getClusterFullName(clusterName, gitRepoName string) (string, error) {
	splitted := strings.Split(gitRepoName, "/")
	var strTmp string
	if len(splitted) > 1 {
		strTmp = splitted[1]
	}
	if strTmp == "" {
		return "", fmt.Errorf("can't set cluster fullname from git repo name. GitRepoName is '%s'", gitRepoName)
	}
	// Prepate string.
	strTmp = strings.ToLower(fmt.Sprintf("%s-%s", clusterName, strTmp))
	// Truncate string.
	strTmp = truncateString(strTmp, 63)
	log.Debugf("ClusterFullName is '%s'", strTmp)

	return strTmp, nil
}

func getManifestsList(path string) ([]string, error) {
	mList, err := filepath.Glob(path + "/*")
	if err != nil {
		return nil, err
	}
	if len(mList) == 0 {
		return nil, fmt.Errorf("no manifest found")
	}
	return mList, nil
}
