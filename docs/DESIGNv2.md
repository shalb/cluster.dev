# Cluster.dev: Cloud infrastructures management tool.

Cluster.dev helps you manage Cloud Native Infrastructures with a simple declarative manifests - Infra Templates.

So you can describe whole infrastructure an deploy it with single tool.

InfraTemplate could contain different technology patterns, like Terraform modules, Shell scripts, Kubernetes manifests, Helm charts, Kustomize and ArgoCD/Flux applications, OPA policies etc..

You can store, test, and distribute your infrastructure pattern as a complete versioned set of technologies.

**Table of contents:**

1) [Concept](#Concept)
2) [Install](#Install)
    - [Prerequisites](#Prerequisites)
    - [Download from release](#Download-from-release)
    - [Build from source](#Build-from-source)
3) [Quick start]()
    - [AWS]()
    - [Google Cloud]()
    - [Azure]()
    - [DigitalOcean]()
4) [Reference]()
    - [Cli commands]()
    - [Cli options]()
5) [Project configuration]()
    - [Project]()
    - [Infrastructures]()
    - [Backends]()
    - [Secrets]()
      - [SOPS]()
      - [Amazon secret manager]()
6) [Template configuration]()
    - [Basics]()
    - [Functions]()
    - [Modules]()
      - [Terraform]()
      - [Helm]()
      - [Kubernetes]()
      - [Printer]()
7) [Generators]()
    - [Project]()
    - [Secret]()

## Concept

### Common Infrastructure Project Structure
```bash
# Common Infrastructure Project Structure
[Project in Git Repo]
  project.yaml           # (Required) Global variables and settings
  [filename].yaml        # (Required at least one) Different project's objects in yaml format (infrastructure, backend etc). 
                         # Se details in configuration reference.
  /templates             # Pre-defined infra patterns. See details in template configuration reference.
    aws-kubernetes.yaml
    cloudflare-dns.yaml
    do-mysql.yaml

    /files               # Some files, used in templates.
      deployment.yaml
      config.cfg
```

### Infrastructure Reconcilation

```bash
# Single command reconcile the whole project
cdev apply
```

Would:

 1. Decode all required secrets.
 2. Template infrastructure variables with global project variables and secrets.
 3. Pull and diff project state and build a dependency graph.
 4. Invoke all required modules in parralel manner.
    ex: `sops decode`, `terraform apply`, `helm install`, etc..

### Demo Video

video will be uploaded later 

## Install 
### Prerequisites
#### Terraform
Cdev client uses the terraform binary. The required terraform version is ~13 or higher. Refer to the [Terraform installation instructions](https://www.terraform.io/downloads.html) to install terraform.

Terraform installation example for linux amd64:
```bash
$ curl -O https://releases.hashicorp.com/terraform/0.14.7/terraform_0.14.7_linux_amd64.zip
$ unzip terraform_0.14.7_linux_amd64.zip
$ mv terraform /usr/local/bin/
```
#### SOPS
For **creating** and **editing** SOPS secrets, cdev uses sops binary. But sops binary in **not required** for decrypting and using sops secrets. All cdev reconcilation processes (build, plan, apply) is not required sops, so no need to install it for pipelines.

See [SOPS installation instructions](https://github.com/mozilla/sops#download) on official repo.

Also see [Secrets section]() in this documentation.

### Download from release

Binaries of the latest stable release are available at [releases page](https://github.com/shalb/cluster.dev/releases). This documentation is suitable for **v0.4.0 or higher**
cluster.dev client. 

Installation example for linux amd64:
```bash
$ wget https://github.com/shalb/cluster.dev/releases/download/v0.4.0-rc1/cluster.dev_v0.4.0-rc1_linux_amd64.tgz
$ tar -xzvf cluster.dev_v0.4.0-rc1_linux_amd64.tgz -C /usr/local/bin

$ cdev --help
```

### Build from source

Go version 16 or higher is required. [Golang installation instructions](https://golang.org/doc/install)

To build cluster.dev client from src:
1)  Clone cluster.dev git repo:
```bash
$ git clone --depth 1 --branch v0.4.0-rc1 https://github.com/shalb/cluster.dev/
```
2) Build binary: 
```bash
$ cd cluster.dev/ && make
```
3) Check cdev and move binary to bin folder: 
```bash
$ ./bin/cdev --help
$ mv ./bin/cdev /usr/local/bin/
```

## Quick start
This guide describes how to quickly deploy infrastructure across different cloud providers. To get started, you need to install the [cdev cli](#Download-from-release) and [required software](#Prerequisites). It is also recommended to install a console client for each cloud provider.

### AWS
#### Authentication
First, you need to configure access to the AWS cloud provider.
There are several ways to do this:

- **Environment variables**: provide your credentials via the `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY`, environment variables, representing your AWS Access Key and AWS Secret Key. You could also ue the `AWS_DEFAULT_REGION` or `AWS_REGION` environment variable, to set region, if needed. Example usage:
```bash
$ export AWS_ACCESS_KEY_ID="MYACCESSKEY"
$ export AWS_SECRET_ACCESS_KEY="MYSECRETKEY"
$ export AWS_DEFAULT_REGION="eu-central-1"
```
- **Shared Credentials File (recommended)**: set up an [AWS configuration file](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-files.html) to specify your credentials. 

Credentials file `~/.aws/credentials` example: 
```bash
[cluster-dev]
aws_access_key_id = MYACCESSKEY
aws_secret_access_key = MYSECRETKEY
```
Config: `~/.aws/config` example:
```bash
[profile cluster-dev]
region = eu-central-1
```
Then export `AWS_PROFILE` environment variable. 
```bash
$ export AWS_PROFILE=cluster-dev
```

#### Install AWS client and check access

See how to install AWS cli in [official installation guide](https://docs.aws.amazon.com/cli/latest/userguide/install-cliv2-linux.html), or use commands from example:
```bash
$ curl "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv2.zip"
$ unzip awscliv2.zip
$ sudo ./aws/install
$ aws s3 ls
```

#### Create s3 bucker for states
To store cluster.dev and terraform states, you should create s3 bucket:
```bash
aws s3 mb s3://cdev-quick-start-states
```
#### DNS Zone:
For the built-in AWS example, you need to define a route53 hosted zone. Options:

1) You already have route53 hosted zone.

2) Create new hosted zone, using [route53 documentation example](https://docs.aws.amazon.com/cli/latest/reference/route53/create-hosted-zone.html#examples).

3) Use "cluster.dev" domain for zone delegation.

### Google cloud
### Azure
### DigitalOcean

## Reference
### Cli commands
### Cli options
## Project configuration
### Project
### Infrastructures
### Backends
### Secrets

There are to way use secrets:
#### sops
Secrets are encoded/decoded with [sops](https://github.com/mozilla/sops) utility which could support AWS KMS, GCP KMS, Azure Key Vault and PGP keys.
How to use: 
1. Use console client cdev to create new secret from scratch:
```bash
cdev secret create sops my_local_secret
```

2. Edit secret and set secret data in `encrypted_data:` section.

3. Use references to secret's data it infrastructure template. 
```yaml
---
name: my-infra
template: "templates/k8s.yaml"
kind: infrastructure
backend: aws-backend
variables:
  region: {{ .secret.my_local_secret.some_key }}
....
```

#### aws ssm
cdev client could use aws ssm as secret storage. 

How to use: 
1. create new secret in aws ssm manager using aws cli or web console. Both data format raw and JSON structure are supported.

2. Use console client cdev to create new secret from scratch:
```bash
cdev secret create aws_ssm my_ssm_secret
```

3. Edit secret and set correct region and ssm_secret name in spec.

4. Use references to secret's data it infrastructure template. 
```yaml
---
name: my-infra
template: "templates/k8s.yaml"
kind: infrastructure
backend: aws-backend
variables:
  region: {{ .secret.my_ssm_secret.some_key }} # if secret is raw data use {{ .secret.my_ssm_secret }}
....
```

To list and edit any secret use commands:
```bash
cdev secret ls
```
and 
```bash
cdev secret edit secret_name
```
## Template configuration
### Basics
### Functions
### Modules
#### Terraform
#### Helm
#### Kubernetes
#### Printer

## Generators
### Project
### Secret


