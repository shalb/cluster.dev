package utils

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/apex/log"
	"github.com/google/go-github/v60/github"
	"github.com/shalb/cluster.dev/pkg/config"
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
	pulledTemplatePath := filepath.Join(targetDir, templateName)
	if IsDir(pulledTemplatePath) {
		log.Debugf("Template is already exists, updating...")
		shell, err := executor.NewExecutor(pulledTemplatePath)
		if err != nil {
			return "", fmt.Errorf("get template: %v", err.Error())
		}
		command := fmt.Sprintf("git pull")
		_, errOutput, err := shell.RunMutely(command)
		if err != nil {
			return pulledTemplatePath, fmt.Errorf("get template: %v\n%v", err.Error(), errOutput)
		}
		return filepath.Join(pulledTemplatePath, parsedGitURL.SubDir), nil
	}
	shell, err := executor.NewExecutor(targetDir)
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

func DiscoverCdevLastRelease() error {
	var (
		client         = github.NewClient(nil)
		ctx            = context.Background()
		org     string = "shalb"
		project string = "cluster.dev"
	)

	latestRelease, _, err := client.Repositories.GetLatestRelease(ctx, org, project)
	if err != nil {
		return err
	}
	config.Version = "v0.9.0"
	curVersion, err := semver.NewVersion(config.Version)
	if err != nil {
		return fmt.Errorf("check failed: %v, current version: %v", err, config.Global.Version)
	}
	reqVerConstraints, err := semver.NewConstraint(*latestRelease.TagName)
	if err != nil {
		return fmt.Errorf("check failed: %v, latest stable release: %v", err, *latestRelease.TagName)
	}
	ok, _ := reqVerConstraints.Validate(curVersion)
	if !ok {
		return fmt.Errorf("the new cdev version is available. Current version: '%v', latest stable release: '%v'. Visit https://docs.cluster.dev/installation-upgrade/ to upgrade", curVersion, *latestRelease.TagName)
	}
	return nil
}
