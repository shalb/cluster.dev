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
3) [Cloud providers](#Cloud-providers)
    - [AWS](#AWS)
    - [Google Cloud](#Google-Cloud)
    - [Azure](#Azure)
    - [DigitalOcean](#DigitalOcean)
4) [Quick start](#Quick-start)
4) [Reference](#Reference)
    - [Cli commands](#Cli-commands)
    - [Cli options](#Cli-options)
5) [Project configuration](#Project-configuration)
    - [Project](#Project)
    - [Infrastructures](#Infrastructures)
    - [Backends](#Backends)
    - [Secrets](#Secrets)
      - [SOPS](#SOPS-secret)
      - [Amazon secret manager](#Amazon-secret-manager)
6) [Template configuration](#Template-configuration)
    - [Basics](#Basics)
    - [Functions](#Functions)
    - [Modules](#Modules)
      - [Terraform](#Terraform-module)
      - [Helm](#Helm-module)
      - [Kubernetes](#Kubernetes-module)
      - [Printer](#Printer-module)

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

Also see [Secrets section](SOPS-secret) in this documentation.

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

## Cloud providers
This guide describes how to set up cloud providers so that `cdev` can start deploying.
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
$ aws s3 mb s3://cdev-states
```
#### DNS Zone:
For the built-in AWS example, you need to define a route53 hosted zone. Options:

1) You already have route53 hosted zone.

2) Create new hosted zone, using [route53 documentation example](https://docs.aws.amazon.com/cli/latest/reference/route53/create-hosted-zone.html#examples).

3) Use "cluster.dev" domain for zone delegation.

### Google cloud
#### Auth:
See [terraform Google cloud provider documentation](https://registry.terraform.io/providers/hashicorp/google/latest/docs/guides/provider_reference#authentication) 
### Azure
#### Auth:
See [terraform Azure provider documentation](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs#authenticating-to-azure)
### DigitalOcean
#### Auth:
Create [an access token](https://www.digitalocean.com/docs/apis-clis/api/create-personal-access-token/). 
Export variable: 
```bash
$ export DIGITALOCEAN_TOKEN="MyToken"
```
How to use DO spaces bucket as backend, see [here](https://www.digitalocean.com/community/questions/spaces-as-terraform-backend)

## Quick start
This guide describes how to quickly create your first project and deploy it. To get started, you need to install the [cdev cli](#Download-from-release) and [required software](#Prerequisites). It is also recommended to install a console client for chosen cloud provider.
1) [Install cdev and soft](#Install).
2) Prepare [cloud provider](#Cloud-providers).
3) Create new empty dir for project and use [cdev generators](#Generators) to create new project from template:
```bash
$ mkdir my-cdev-project && cd my-cdev-project
$ cdev new project
```
4) Choose one of the available projects. Check out the description of the example. Enter the data required for the generator.
5) After finishing work with the generator - check the project:
```bash
$ cdev project info
```
6) Edit [project](#Project-configuration) and [template](#Template-configuration) configuration, if needed. 
```bash
$ vim project.yaml
$ vim infra.yaml
$ vim templates/aws-k3s.yaml # (name depends of chosen option in step 4)
```
7) Apply the project:
```bash
$ cdev apply -l debug
```

## Reference
### Cli commands
Available Commands:
  - `build`       Build all modules - read project configuration. Check it, render templates and generate code of all modules in tmp dir: `./cluster.dev/project-name/`
  - `apply`       Apply all modules - build project (see build command), calculate dependencies, create and check graph. Deploy all modules according to the graph.
  - `destroy`     Destroy all modules - build project (see build command), calculate dependencies, create and check the reverse graph. Run destroy scenario for all modules according to the graph.
  - `help`        Help about any command
  - `new`         Code generator. Creates new 'project' or 'secret' from template in interactive mode.
  - `plan`        Plan all modules - build project. Try to run the plan scenario for modules. Modules often refer to the remote states of other modules. Because of this, the plan command may fail if the remote state does not already exist.
  - `project`     Manage projects:
    - `info`      Read project and info message.
  - `secret`      Manage secrets:
    - `ls`        List secrets in current project.
    - `edit`      Edit secret by name. Usage: `cdev secret edit secret-name`
### Cli options
Available global flags:
  - `--cache`             Use previously cached build directory
  - `-l, --log-level string`   Set the logging level ('debug'|'info'|'warn'|'error'|'fatal') (default "info")
  - `--parallelism int`    Max parallel threads for module applying (default - `3`)
  - `--trace`              Print functions trace info in logs (Mainly used for development)

## Project configuration
`cdev` reads configuration from current directory. Reads all files by mask: `*.yaml`. It is allowed to place several yaml configuration objects in one file, separating them "---". The exception is the project.yaml configuration and files with secrets.

Project represents the single scope for infrastructures unde what they are stored and reconciled. The dependencies between different infrastructures could be used under the project scope. Project could host a global variables that could be accessed to template target infrastructure.

### Project
File: `project.yaml`. Required. 
Contain global project variables, which can be used in other configuration objects such as backend or infrastructure (except `secrets`). File `project.conf` is not renders with template, you cannot use template units in it. 

Example `project.yaml`:
```yaml
name: my_project
kind: project
variables:
  organization: shalb
  region: eu-central-1
  state_bucket_name: cdev-states
```
- `name`: project name. *Required*.
- `kind`: object kind. Must be `project`. *Required*.
- `variables`: a set of data in yaml format that can be referenced in other configuration objects. For the example above, the link to the name of the organization will look like this: `{{ .project.variables.organization }}`
### Infrastructures
File: searching in `./*.yaml`. *Required at least one*.
Infrastructure object (`kind: infrastructure`) contain reference to template, variables to render that template and backend for states. 

Example: 
```yaml
# Define infrastructure itself
name: k3s-infra
template: "templates/aws-k3s.yaml"
kind: infrastructure
backend: aws-backend
variables:
  bucket: {{ .project.variables.state_bucket_name }} # Using project variables.
  region: {{ .project.variables.region }}
  organization: {{ .project.variables.organization }}
  domain: cluster.dev
  instance_type: "t3.medium"
  vpc_id: "vpc-5ecf1234"
```
- `name`: infrastructure name. *Required*.
- `kind`: object kind. `infrastructure`. *Required*.
- `backend`: name of [backend](#Backends), which will be used to store the states of this infrastructure. *Required*
- `variables`: data set for [template rendering](#Template-configuration). 

### Backends
File: searching in `./*.yaml`. *Required at least one*.
An object that describes a backend storage for terraform and cdev states.
In backends configuration you can use any options of appropriate terraform backend. They will be converted as is.
Currently 4 types of backends are supported:

- `s3` AWS S3 backend:
```yaml
name: aws-backend
kind: backend
provider: s3
spec:
  bucket: cdev-states
  region: {{ .project.variables.region }}
```
- `do` DigitalOcean spaces backend:
```yaml
name: do-backend
kind: backend
provider: do
spec:
  bucket: cdev-states
  region: {{ .project.variables.region }}
  access_key: {{ env "SPACES_ACCESS_KEY_ID" }}
  secret_key: {{ env "SPACES_SECRET_ACCESS_KEY" }}
```
- `azurerm` Microsoft azurem:
```yaml
name: gcs-b
kind: backend
provider: azurerm
spec:
  resource_group_name: "StorageAccount-ResourceGroup"
  storage_account_name: "example"
  container_name: "cdev-states"
```
- `gcs` Google Cloud backend:
```yaml
name: do-backend
kind: backend
provider: gcs
spec:
  bucket: cdev-states
  prefix: pref
```

### Secrets

There are to way use secrets:
#### SOPS secret
Secrets are encoded/decoded with [sops](https://github.com/mozilla/sops) utility which could support AWS KMS, GCP KMS, Azure Key Vault and PGP keys.
How to use: 
1. Use console client cdev to create new secret from scratch:
```bash
$ cdev secret create sops my_local_secret
```
2. Use interactive menu to create secret.
3. Edit secret and set secret data in `encrypted_data:` section.
4. Use references to secret's data it infrastructure template (examples you can find in generated secret file). 


#### Amazon secret manager
cdev client could use aws ssm as secret storage. 

How to use: 
1. create new secret in aws secret manager using aws cli or web console. Both data format raw and JSON structure are supported.

2. Use console client cdev to create new secret from scratch:
```bash
$ cdev secret create
```
3. Answer the questions. For `Name of secret in AWS Secrets manager` enter name of aws secret, created above.
4. Use references to secret's data it infrastructure template (examples you can find in generated secret file). 


To list and edit any secret use commands:
```bash
$ cdev secret ls
```
and 
```bash
$ cdev secret edit secret_name
```
## Template configuration
### Basics
Template represents yaml structure with array of different invocation [modules](#Modules)
Common view:
```yaml
modules: 
  - module1
  - module2
  - module3
  ...
```

Template could utilize all kind of go-template and sprig functions (similar to Helm). Along that it is enhanced with functions like insertYAML that could pass yaml blocks directly.

### Functions
1) [Base go-template language functions](https://golang.org/pkg/text/template/#hdr-Functions).
2) [Sprig functions](https://masterminds.github.io/sprig/).
3) Enhanced functions: all functions described above allow you to modify the template text. Besides these, some special enhanced functions are available. They may not be used everywhere. They are integrated with the functionality of the program and with the yaml syntax:
  - `insertYAML` - pass yaml block as value of target yaml template. **Argument**: data to pass, any value or reference to block.
  **Allowed use**: only as full yaml value, in module `inputs`. Example:
source yaml:
```yaml
values:
  node_groups:
    - name: ng1
      min_size: 1
      max_size: 5
    - name: ng2
      max_size: 2
      type: spot
```
target yaml template:
```yaml
modules: 
  - name: k3s
    type: terraform
    node_groups: {{ insertYAML .values.node_groups }}
```
rendered template:
```yaml
modules: 
  - name: k3s
    type: terraform
    node_groups:     
    - name: ng1
      min_size: 1
      max_size: 5
    - name: ng2
      max_size: 2
      type: spot
```
  - `remoteState` - used for passing data between modules and infrastructures. **Argument**: string, path to remote state consisting of 3 parts separated by a dot: `"infra_name.module_name.output_name"`. Since the name of the infrastructure is unknown inside the template, you can use "this" instead:`"this.module_name.output_name"`. **Allowed use**: as yaml value , only in module `inputs`.
 

### Modules
All modules described below have a common format and common fields. Base example: 
```yaml
  - name: k3s
    type: terraform
    depends_on: 
      - this.module1_name
      - this.module2_name
#   depends_on: this.module1_name # is allowed to use string for single, or list for multiple dependencies
    pre_hook:
      command: "echo pre_hook"
      # script: "./scripts/hook.sh"
      on_apply: true
      on_destroy: false
      on_plan: false
    post_hook:
      # command: "echo post_hook" 
      script: "./scripts/hook.sh"
      on_apply: true
      on_destroy: false
      on_plan: false
```
- `name` - module name. *Required*.
- `type` - module type. One of: `terraform`, `helm`, `kubernetes`, `printer`. See below.
- `depends_on` - *string* or *list of strings*. One or multiple dependencies of module in format "infra_name.module_name". Since the name of the infrastructure is unknown inside the template, you can use "this" instead:`"this.module_name.output_name"` 
- `pre_hook` and `post_hook` blocks: describes the shell commands to be executed before and after the module, respectively. The commands will be executed in the same context as the actions of the module. All environment variables will be available between them.
    - `command` - *string*. Shell command in text format. Will be executed in bash -c "command". Can be used if the "script" option is not used. One of `command` or `script` is required.
    - `script` - *string* path to shell script file which is relative to template directory. Can be used if the "command" option is not used. One of `command` or `script` is required.
    - `on_apply` *bool*, *optional* turn off/on when module applying. **Default: "true"**
    - `on_destroy` - *bool*, *optional* turn off/on when module destroying. **Default: "false"**
    - `on_plan` - *bool*, *optional* turn off/on when module plan executing. **Default: "false"**

#### Terraform module
Describes direct Terraform module invocation. 

Example: 
```yaml
modules:
  - name: vpc
    type: terraform
    version: "2.77.0"
    source: terraform-aws-modules/vpc/aws
    inputs:
      name: {{ .name }}
      azs: {{ insertYAML .variables.azs }}
      vpc_id: {{ .variables.vpc_id }}
```
In addition to common options, the following are available:
- `source` - *string*, *required*. Terraform module [source](https://www.terraform.io/docs/language/modules/syntax.html#source). **It is not allowed to use local folders in source!** 
- `version` - *string*, *optional*. Module [version](https://www.terraform.io/docs/language/modules/syntax.html#version).
- `inputs` - *map of any*, *required*. Map, which correspond to [input variables](https://www.terraform.io/docs/language/values/variables.html) defined by the module. This block allows to use functions `remoteState` and `insertYAML`
#### Helm module
Describes [Terraform helm provider](https://registry.terraform.io/providers/hashicorp/helm/latest/docs) invocation.

Example: 
```yaml
modules:
  - name: argocd
    type: helm
    source:
      repository: "https://argoproj.github.io/argo-helm"
      chart: "argo-cd"
      version: "2.11.0"
    kubeconfig: ../kubeconfig
    depends_on: this.k3s
    pre_hook:
      script: ./scripts/get_kubeconfig.sh ./kubeconfig
      on_destroy: true
      on_plan: true
    additional_options:
      namespace: "argocd"
      create_namespace: true
    inputs:
      global.image.tag: v1.8.3
      service.type: LoadBalancer
```
In addition to common options, the following are available:
- `source` - *map*, *required*. Block describes helm chart source. 
    - `chart`, `repository`, `version` - correspond to options of the same name of helm_release resource. See [chart](https://registry.terraform.io/providers/hashicorp/helm/latest/docs/resources/release#chart), [repository](https://registry.terraform.io/providers/hashicorp/helm/latest/docs/resources/release#repository) and [version](https://registry.terraform.io/providers/hashicorp/helm/latest/docs/resources/release#version).
    - `kubeconfig` - *string*, *required*. Path to kubeconfig file which is relative to directory where the module was executed.
    - `additional_options` - *map of any*, *optional*. Corresponds to [helm_release resource options](https://registry.terraform.io/providers/hashicorp/helm/latest/docs/resources/release#argument-reference). Will be pass as is.
    - `inputs` - *map of any*, *optional*. Map, which represents [helm release sets](https://registry.terraform.io/providers/hashicorp/helm/latest/docs/resources/release#set). This block allows to use functions `remoteState` and `insertYAML`.
    
    For example: 
    ```yaml
    inputs:
      global.image.tag: v1.8.3
      service.type: LoadBalancer
    ```
    corresponds to 
    ```hcl
    set {
      name = "global.image.tag"
      value = "v1.8.3"
    }
    set  {
      name = "service.type"
      value = "LoadBalancer"
    }
    ```
#### Kubernetes module
Describes [Terraform kubernetes-alpha provider](https://github.com/hashicorp/terraform-provider-kubernetes-alpha) invocation.

Example: 
```yaml
modules:
  - name: argocd_apps
    type: kubernetes
    source: ./argocd-apps/app1.yaml
    kubeconfig: ../kubeconfig
    depends_on: this.argocd
```

- `source` - *string*, *required*. Path to kubernetes manifest, which will be converted into a representation of kubernetes-alpha provider. **Source file will be rendered with template, and also allows to use of functions `remoteState` and `insertYAML`**
- `kubeconfig` - *string*, *required*. Path to kubeconfig file which is relative to directory where the module was executed.
#### Printer module 
Module is mainly used to see the outputs of other modules in the console logs.

Example:
```yaml
modules:
  - name: print_outputs
    type: printer
    inputs:
      cluster_name: {{ .name }}
      worker_iam_role_arn: {{ remoteState "this.eks.worker_iam_role_arn" }}
```
`inputs` - *any*, *requited* - map, represents data to be printed in the log. This block **allows to use of functions `remoteState` and `insertYAML`**

