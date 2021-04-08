package utils

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/apex/log"
	"github.com/shalb/cluster.dev/pkg/executor"
)

type gitRepo struct {
	URL     string
	Subdir  string
	Version string
}

func GetTemplate(gitURL, targetDir, templateName string) (string, error) {
	log.Debugf("Cloning template from repo: %v, %v, %v", gitURL, targetDir, templateName)
	parsedGitURL, err := parseGitUrl(gitURL)
	if err != nil {
		return "", fmt.Errorf("get template: %v", err.Error())
	}
	pulledTempplatePath := filepath.Join(targetDir, templateName)
	if IsDir(pulledTempplatePath) {
		log.Debugf("Template is already exists, updating...")
		shell, err := executor.NewBashRunner(pulledTempplatePath)
		if err != nil {
			return "", fmt.Errorf("get template: %v", err.Error())
		}
		command := fmt.Sprintf("git pull")
		_, errOutput, err := shell.RunMutely(command)
		if err != nil {
			return pulledTempplatePath, fmt.Errorf("get template: %v\n%v", err.Error(), errOutput)
		}
		return pulledTempplatePath, nil
	}
	shell, err := executor.NewBashRunner(targetDir)
	if err != nil {
		return "", fmt.Errorf("get template: %v", err.Error())
	}
	command := fmt.Sprintf("git clone --single-branch --depth=1 ")
	if parsedGitURL.Version != "" {
		command = command + "-b " + parsedGitURL.Version + " "
	}
	command = command + parsedGitURL.URL + " " + templateName
	_, errOutput, err := shell.RunMutely(command)
	if err != nil {
		return "", fmt.Errorf("get template: %v\n %v", err.Error(), errOutput)
	}
	return filepath.Join(pulledTempplatePath, parsedGitURL.Subdir), nil
}

func parseGitUrl(gitURL string) (repo gitRepo, err error) {
	res := strings.Split(gitURL, "?ref=")
	var url string
	if len(res) == 2 {
		repo.Version = res[1]
	}
	url = res[0]
	splittedSubdir := strings.Split(url, "//")
	if strings.HasPrefix(url, "https://") {
		if len(splittedSubdir) > 2 {
			repo.Subdir = splittedSubdir[2]
		}
		repo.URL = splittedSubdir[0] + "//" + splittedSubdir[1]
	} else if strings.HasPrefix(url, "git@") {
		if len(splittedSubdir) > 1 {
			repo.Subdir = splittedSubdir[1]
		}
		repo.URL = splittedSubdir[0]
	} else {
		err = fmt.Errorf("parse git url: bad url '%v'", gitURL)
		return
	}
	return
}
