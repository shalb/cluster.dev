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
