# Installer

Implement [cli-installer specification](../../docs/design/cli-installer-design.md)

**Status: `In progress`**

## Usage

Build:

```bash
PATH_TO_INSTALLER="/full_path_to_repo/install/installer/"
PATH_TO_INSTALLER=/home/vm/code/SHALB/cluster.dev/cluster.dev/install/installer
docker build -t cli-installer "$PATH_TO_INSTALLER"
```

Run:

```bash
docker run \
    -v "$(pwd)":/app/current_dir \
    -e GITHUB_TOKEN \
    cli-installer \
    install
```

`-e` - Mount host env vars inside container


#### Main features support:

##### 1. Create or reuse infrastructure repo within git hosting provider

Supported git-repo types:

- [x] Cloned (w/ origins) empty repo
- [x] Cloned (w/ origins) non-empty repo
- [ ] `WIP` Created (w/o origins) empty repo
- [ ] `WIP` Created (w/o origins) non-empty repo


##### 2. Select and create cloud user and required permissions

- [x] Select cloud
- [x] Get all needed creds for cloud user (AWS)
- [x] Create cloud user and required permissions (AWS)
- [x] Get creds for new user (AWS)

##### 3. Git an Git Providers

- [x] Add secrets to github repo
- [ ] Add secrets to gitlab repo

##### 4. Populate repo with files and push

- [x] Populate repo with sample files
- [ ] Commit and push all code.
- [ ] Set `installed: true` and push - Install the cluster
- [ ] Open Edit View for the first cluster config

##### 6. Display credentials with output

- [ ] Supported

##### 7. Create binary

- [ ] [Creating standalone Mac OS X applications](https://www.metachris.com/2015/11/create-standalone-mac-os-x-applications-with-python-and-py2app/) | [py2app docs](https://py2app.readthedocs.io/en/latest/)
- [ ] [py2exe - Create standalone Windows app](https://www.py2exe.org/)



## Style guide

Code style checked by autopep8 and pylint

See its settings in [`.pre-commit-config.yaml`](https://github.com/shalb/cluster.dev/blob/master/.pre-commit-config.yaml)

Function documentation style based on [google style example](https://sphinxcontrib-napoleon.readthedocs.io/en/latest/example_google.html).

## Local debug and run

### Build

```bash
docker-compose build
```

### Run

```bash
docker-compose run app
```

### Run tests

Non-iteractive:

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
