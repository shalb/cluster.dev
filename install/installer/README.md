# Installer

Implement [cli-installer specification](../../docs/design/cli-installer-design.md)

**Status: `In progress`**

* [Usage](#usage)
* [Build and push new image](#build-and-push-new-image)
  * [Main features support](#main-features-support)
    * [1. Create or reuse infrastructure repo within git hosting provider](#1-create-or-reuse-infrastructure-repo-within-git-hosting-provider)
    * [2. Select and create cloud user and required permissions](#2-select-and-create-cloud-user-and-required-permissions)
    * [3. Git an Git Providers](#3-git-an-git-providers)
    * [4. Populate repo with files and push](#4-populate-repo-with-files-and-push)
    * [6. Display credentials with output](#6-display-credentials-with-output)
    * [7. Create binary](#7-create-binary)
* [Style guide](#style-guide)
* [Local debug and run](#local-debug-and-run)
  * [Requirements](#requirements)
  * [Build](#build)
  * [Run](#run)
  * [Run tests](#run-tests)
* [Useful links](#useful-links)


## Usage

1. Check that your AWS user:

    1.1. [Have this rights](../installer_aws_install_req_permissions.json)

    1.2. You have it programmatic access key and secret.
2. Create github repo and clone it.
3. Then, run inside cloned repo:

```bash
TAG=0.1.3
docker run -it \
    -v "$(pwd)":/app/current_dir \
    -v "$HOME"/.gitconfig:/home/cluster.dev/.gitconfig:ro \
    -v "$HOME"/.ssh:/home/cluster.dev/.ssh:ro \
    -v "$HOME"/.aws:/home/cluster.dev/.aws:ro \
    -e GITHUB_TOKEN \
    -e UID=$(id -u) \
    -e GID=$(id -g) \
    shalb/cluster.dev-cli-installer:$TAG \
    install
```

`-e` - Mount host env vars inside container


## Build and push new image

```bash
TAG=0.1.3
PATH_TO_INSTALLER="/full_path_to_repo/install/installer/"

docker build -t shalb/cluster.dev-cli-installer:latest -t shalb/cluster.dev-cli-installer:$TAG "$PATH_TO_INSTALLER"
docker push shalb/cluster.dev-cli-installer:latest
docker push shalb/cluster.dev-cli-installer:$TAG
```

### Main features support

#### 1. Create or reuse infrastructure repo within git hosting provider

Supported git-repo types:

- [x] Cloned (w/ origins) empty repo
- [x] Cloned (w/ origins) non-empty repo
- [ ] `WIP` Created (w/o origins) empty repo
- [ ] `WIP` Created (w/o origins) non-empty repo


#### 2. Select and create cloud user and required permissions

- [x] Select cloud
- [x] Get all needed creds for cloud user (AWS)
- [x] Create cloud user and required permissions (AWS)
- [x] Get creds for new user (AWS)

#### 3. Git an Git Providers

- [x] Add secrets to github repo
- [ ] Add secrets to gitlab repo

#### 4. Populate repo with files and push

- [x] Populate repo with sample files
- [x] Commit and push all code.
- [x] Set `installed: true`
- [x] Open Edit View for the first cluster config
- [x] Commit and push - Install the cluster

#### 6. Display credentials with output

- [ ] Supported

#### 7. Create binary

- [ ] [Creating standalone Mac OS X applications](https://www.metachris.com/2015/11/create-standalone-mac-os-x-applications-with-python-and-py2app/) | [py2app docs](https://py2app.readthedocs.io/en/latest/)
- [ ] [py2exe - Create standalone Windows app](https://www.py2exe.org/)



## Style guide

Code style checked by autopep8 and pylint

See its settings in [`.pre-commit-config.yaml`](https://github.com/shalb/cluster.dev/blob/master/.pre-commit-config.yaml)

Function documentation style based on [google style example](https://sphinxcontrib-napoleon.readthedocs.io/en/latest/example_google.html).

## Local debug and run

### Requirements

* Docker 19.03+
* docker-compose 1.25+
* Python 3.8+ (for `pre-commit`: `mypy` and `pylint` checks)

### Build

```bash
docker-compose build
```

### Run

```bash
docker-compose run app
```

### Run tests

Non-interactive:

```bash
docker-compose run tests
```

Interactive:

```bash
docker-compose run mount_only_current_dir
docker-compose run without_repo
docker-compose run empty_repo
docker-compose run non_empty_repo
docker-compose run cloned_ssh_empty_repo
docker-compose run cloned_https_empty_repo
```


## Useful links

* [PyInquirer (interactive part) examples](https://github.com/CITGuru/PyInquirer#examples)
* [GitPython Tutorial (for creation repo for user)](https://gitpython.readthedocs.io/en/stable/tutorial.html)
* [Creating standalone Mac OS X applications](https://www.metachris.com/2015/11/create-standalone-mac-os-x-applications-with-python-and-py2app/) | [py2app docs](https://py2app.readthedocs.io/en/latest/)
* [py2exe - Create standalone Windows app](https://www.py2exe.org/)
