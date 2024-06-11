package utils

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/apex/log"
	"github.com/shalb/cluster.dev/pkg/executor"
)

type GitRepo struct {
	URL        string
	RepoString string // Repo url without the protocol and the subdirectory e.g. github.com/org/repo?ref=tag .
	SubDir     string
	Version    string
}

func GetTemplate(gitURL, targetDir, templateName string) (string, error) {
	log.Debugf("Cloning template from repo: %v, %v, %v", gitURL, targetDir, templateName)
	parsedGitURL, err := ParseGitUrl(gitURL)
	if err != nil {
		return "", fmt.Errorf("get template: %v", err.Error())
	}
	var interruptMoc = false
	pulledTemplatePath := filepath.Join(targetDir, templateName)
	if IsDir(pulledTemplatePath) {
		log.Debugf("Template is already exists, updating...")
		shell, err := executor.NewExecutor(pulledTemplatePath, &interruptMoc)
		if err != nil {
			return "", fmt.Errorf("get template: %v", err.Error())
		}
		command := "git pull"
		_, errOutput, err := shell.RunMutely(command)
		if err != nil {
			return pulledTemplatePath, fmt.Errorf("get template: %v\n%v", err.Error(), errOutput)
		}
		return filepath.Join(pulledTemplatePath, parsedGitURL.SubDir), nil
	}
	shell, err := executor.NewExecutor(targetDir, &interruptMoc)
	if err != nil {
		return "", fmt.Errorf("get template: %v", err.Error())
	}
	command := "git clone --single-branch --depth=1 "
	if parsedGitURL.Version != "" {
		command = command + "-b " + parsedGitURL.Version + " "
	}
	command = command + parsedGitURL.URL + " " + templateName
	_, errOutput, err := shell.RunMutely(command)
	if err != nil {
		return "", fmt.Errorf("get template: %v\n %v", err.Error(), errOutput)
	}
	return filepath.Join(pulledTemplatePath, parsedGitURL.SubDir), nil
}

func ParseGitUrl(gitURL string) (repo GitRepo, err error) {
	res := strings.Split(gitURL, "?ref=")
	var url string
	if len(res) == 2 {
		repo.Version = res[1]
	}
	url = res[0]
	splittedSubdir := strings.Split(url, "//")
	if strings.HasPrefix(url, "https://") {
		if len(splittedSubdir) > 2 {
			repo.SubDir = splittedSubdir[2]
		}
		repo.URL = splittedSubdir[0] + "//" + splittedSubdir[1]
		repo.RepoString = splittedSubdir[1] + "?ref=" + repo.Version
	} else if strings.HasPrefix(url, "git@") {
		if len(splittedSubdir) > 1 {
			repo.SubDir = splittedSubdir[1]
		}
		repo.URL = splittedSubdir[0]
		repo.RepoString = splittedSubdir[0] + "?ref=" + repo.Version
	} else {
		err = fmt.Errorf("parse git url: bad url '%v'", gitURL)
		return
	}
	// log.Warnf("ParseGitUrl %+v", repo)
	return
}
