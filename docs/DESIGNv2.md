# Cluster.dev: Cloud infrastructures' management tool

Cluster.dev helps you manage Cloud Native Infrastructures with simple declarative manifests - Infra Templates.

So you can describe a whole infrastructure and deploy it with a single tool.

Infra Template could contain different technology patterns, such as Terraform modules, Shell scripts, Kubernetes manifests, Helm charts, Kustomize and ArgoCD/Flux applications, OPA policies etc.

You can store, test, and distribute your infrastructure pattern as a complete versioned set of technologies.

**Table of contents:**

* [Concept](#concept)
  * [Common Infrastructure Project Structure](#common-infrastructure-project-structure)
  * [Infrastructure Reconcilation](#infrastructure-reconcilation)
  * [Demo Video](#demo-video)
* [Install](#install)
  * [Prerequisites](#prerequisites)
    * [Terraform](#terraform)
    * [SOPS](#sops)
  * [Download from release](#download-from-release)
  * [Build from source](#build-from-source)
* [Cloud providers](#cloud-providers)
  * [AWS](#aws)
    * [Authentication](#authentication)
    * [Install AWS client and check access](#install-aws-client-and-check-access)
    * [Create S3 bucket for states](#create-s3-bucket-for-states)
    * [DNS Zone](#dns-zone)
  * [Google cloud](#google-cloud)
    * [Auth](#auth)
  * [Azure Authentication](#azure-authentication)
  * [DigitalOcean Authentication](#digitalocean-authentication)
* [Quick start](#quick-start)
* [Reference](#reference)
  * [Cli commands](#cli-commands)
  * [Cli options](#cli-options)
* [Project configuration](#project-configuration)
  * [Project](#project)
  * [Infrastructures](#infrastructures)
  * [Backends](#backends)
  * [Secrets](#secrets)
    * [SOPS secret](#sops-secret)
    * [Amazon secret manager](#amazon-secret-manager)
* [Template configuration](#template-configuration)
  * [Basics](#basics)
  * [Functions](#functions)
  * [Modules](#modules)
    * [Terraform module](#terraform-module)
    * [Helm module](#helm-module)
    * [Kubernetes module](#kubernetes-module)
    * [Printer module](#printer-module)

## Concept

### Common Infrastructure Project Structure

```bash
# Common Infrastructure Project Structure
[Project in Git Repo]
  project.yaml           # (Required) Global variables and settings
  [filename].yaml        # (Required at least one) Different project's objects in yaml format (infrastructure, backend etc).
                         # See details in configuration reference.
  /templates             # Pre-defined infra patterns. See details in template configuration reference.
    aws-kubernetes.yaml
    cloudflare-dns.yaml
    do-mysql.yaml

    /files               # Some files used in templates.
      deployment.yaml
      config.cfg
```

### Infrastructure Reconcilation

```bash
# Single command reconciles the whole project
cdev apply
```

Running the command will:

 1. Decode all required secrets.
 2. Template infrastructure variables with global project variables and secrets.
 3. Pull and diff project state and build a dependency graph.
 4. Invoke all required modules in a parralel manner.
    ex: `sops decode`, `terraform apply`, `helm install`, etc.

### Demo Video

video will be uploaded soon.

## Install

### Prerequisites

#### Terraform

Cdev client uses the Terraform binary. The required Terraform version is ~13 or higher. Refer to the [Terraform installation instructions](https://www.terraform.io/downloads.html) to install Terraform.

Terraform installation example for linux amd64:

```bash
curl -O https://releases.hashicorp.com/terraform/0.14.7/terraform_0.14.7_linux_amd64.zip
unzip terraform_0.14.7_linux_amd64.zip
mv terraform /usr/local/bin/
```

#### SOPS

For **creating** and **editing** SOPS secrets, cdev uses SOPS binary. But the SOPS binary is **not required** for decrypting and using SOPS secrets. As none of cdev reconcilation processes (build, plan, apply) requires SOPS to be performed, you don't have to install it for pipelines.

See [SOPS installation instructions](https://github.com/mozilla/sops#download) in official repo.

Also see [Secrets section](#sops-secret) in this documentation.

### Download from release

Binaries of the latest stable release are available on the [releases page](https://github.com/shalb/cluster.dev/releases). This documentation is suitable for **v0.4.0 or higher**
cluster.dev client.

Installation example for linux amd64:

```bash
wget https://github.com/shalb/cluster.dev/releases/download/v0.4.0-rc1/cluster.dev_v0.4.0-rc1_linux_amd64.tgz
tar -xzvf cluster.dev_v0.4.0-rc1_linux_amd64.tgz -C /usr/local/bin

cdev --help
```

### Build from source

Go version 16 or higher is required. [Golang installation instructions](https://golang.org/doc/install)

To build cluster.dev client from src:

Clone cluster.dev git repo:

```bash
git clone --depth 1 --branch v0.4.0-rc1 https://github.com/shalb/cluster.dev/
```

Build binary:

```bash
cd cluster.dev/ && make
```

Check cdev and move binary to bin folder:

```bash
./bin/cdev --help
mv ./bin/cdev /usr/local/bin/
```

## Cloud providers

This section contains guidelines on cloud settings required for `cdev` to start deploying.

### AWS

#### Authentication

First, you need to configure access to the AWS cloud provider.
There are several ways to do this:

* **Environment variables**: provide your credentials via the `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY`, the environment variables that represent your AWS Access Key and AWS Secret Key. You can also use the `AWS_DEFAULT_REGION` or `AWS_REGION` environment variable to set region, if needed. Example usage:

```bash
export AWS_ACCESS_KEY_ID="MYACCESSKEY"
export AWS_SECRET_ACCESS_KEY="MYSECRETKEY"
export AWS_DEFAULT_REGION="eu-central-1"
```

* **Shared Credentials File (recommended)**: set up an [AWS configuration file](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-files.html) to specify your credentials.

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
export AWS_PROFILE=cluster-dev
```

#### Install AWS client and check access

See how to install AWS cli in [official installation guide](https://docs.aws.amazon.com/cli/latest/userguide/install-cliv2-linux.html), or use commands from example:

```bash
curl "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv2.zip"
unzip awscliv2.zip
sudo ./aws/install
aws s3 ls
```

#### Create S3 bucket for states

To store cluster.dev and Terraform states, you should create an S3 bucket:

```bash
aws s3 mb s3://cdev-states
```

#### DNS Zone

For the built-in AWS example, you need to define a Route 53 hosted zone. Options:

1) You already have a Route 53 hosted zone.

2) Create a new hosted zone using a [Route 53 documentation example](https://docs.aws.amazon.com/cli/latest/reference/route53/create-hosted-zone.html#examples).

3) Use "cluster.dev" domain for zone delegation.

### Google cloud

#### Auth

See [Terraform Google cloud provider documentation](https://registry.terraform.io/providers/hashicorp/google/latest/docs/guides/provider_reference#authentication)

### Azure Authentication

See [Terraform Azure provider documentation](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs#authenticating-to-azure)

### DigitalOcean Authentication

Create [an access token](https://www.digitalocean.com/docs/apis-clis/api/create-personal-access-token/).
Export variable:

```bash
export DIGITALOCEAN_TOKEN="MyToken"
```

For details on using DO spaces bucket as a backend, see [here](https://www.digitalocean.com/community/questions/spaces-as-terraform-backend)

## Quick start

This guide describes how to quickly create your first project and deploy it. To get started, you need to install the [cdev cli](#Download-from-release) and [required software](#Prerequisites). It is also recommended to install a console client for the chosen cloud provider.

1. [Install cdev and required software](#Install).

2. Prepare [cloud provider](#Cloud-providers).

3. Create a new empty dir for the project and use [cdev generators](#Generators) to create the new project from a template:

```bash
mkdir my-cdev-project && cd my-cdev-project
cdev new project
```

4. Choose one from the available projects. Check out the description of the example. Enter the data required for the generator.

5. After finished working with the generator, check the project:

```bash
cdev project info
```

6. Edit [project](#Project-configuration) and [template](#Template-configuration) configuration, if needed.

```bash
vim project.yaml
vim infra.yaml
vim templates/aws-k3s.yaml # (the name depends on chosen option in step 4)
```

7. Apply the project:

```bash
cdev apply -l debug
```

## Reference

### Cli commands

Available Commands:

* `build`       Build all modules - read project configuration. Check it, render templates and generate code of all modules in tmp dir: `./cluster.dev/project-name/`

* `apply`       Apply all modules - build project (see build command), calculate dependencies, create and check graph. Deploy all modules according to the graph.

* `destroy`     Destroy all modules - build project (see build command), calculate dependencies, create and check the reverse graph. Run destroy scenario for all modules according to the graph.

* `help`        Help about any command.

* `plan`        Plan all modules - build project. Try to run the plan scenario for modules. Modules often refer to the remote states of other modules. Because of this, the plan command may fail if the remote state does not already exist.

* `project`     Manage projects:

  * `info`      Read project and info message.
  * `create`    Creates new 'project' from template in interactive mode.

* `secret`      Manage secrets:

  * `ls`        List secrets in current project.
  * `edit`      Edit secret by name. Usage: `cdev secret edit secret-name`.
  * `create`    Creates new 'secret' from template in interactive mode.

### Cli options

Available global flags:

* `--cache`             Use previously cached build directory.

* `-l, --log-level string`   Set the logging level ('debug'|'info'|'warn'|'error'|'fatal') (default "info").

* `--parallelism int`    Max parallel threads for module applying (default - `3`).

* `--trace`              Print functions trace info in logs (Mainly used for development).

## Project configuration

`cdev` reads configuration from current directory, i.e. all files by mask: `*.yaml`. It is allowed to place several yaml configuration objects in one file, separating them with "---". The exception is the project.yaml configuration file and files with secrets.

Project represents a single scope for infrastructures within which they are stored and reconciled. The dependencies between different infrastructures can be used within the project scope. Project can host global variables that can be used to template target infrastructure.

### Project

File: `project.yaml`. *Required*.
Contains global project variables that can be used in other configuration objects, such as backend or infrastructure (except `secrets`). Note that the `project.conf` file is not rendered with the template and you cannot use template units in it.

Example `project.yaml`:

```yaml
name: my_project
kind: project
variables:
  organization: shalb
  region: eu-central-1
  state_bucket_name: cdev-states
```

* `name`: project name. *Required*.

* `kind`: object kind. Must be `project`. *Required*.

* `variables`: a set of data in yaml format that can be referenced in other configuration objects. For the example above, the link to the organization name will look like this: `{{ .project.variables.organization }}`.

### Infrastructures

File: searching in `./*.yaml`. *Required at least one*.
Infrastructure object (`kind: infrastructure`) contains reference to a template, variables to render the template and backend for states.

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

* `name`: infrastructure name. *Required*.

* `kind`: object kind. `infrastructure`. *Required*.

* `backend`: name of the [backend](#Backends) that will be used to store the states of this infrastructure. *Required*.

* `variables`: data set for [template rendering](#Template-configuration).

### Backends

File: searching in `./*.yaml`. *Required at least one*.
An object that describes a backend storage for Terraform and cdev states.
In backends' configuration you can use any options of appropriate Terraform backend. They will be converted as is.
Currently 4 types of backends are supported:

* `s3` AWS S3 backend:

```yaml
name: aws-backend
kind: backend
provider: s3
spec:
  bucket: cdev-states
  region: {{ .project.variables.region }}
```

* `do` DigitalOcean spaces backend:

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

* `azurerm` Microsoft azurem:

```yaml
name: gcs-b
kind: backend
provider: azurerm
spec:
  resource_group_name: "StorageAccount-ResourceGroup"
  storage_account_name: "example"
  container_name: "cdev-states"
```

* `gcs` Google Cloud backend:

```yaml
name: do-backend
kind: backend
provider: gcs
spec:
  bucket: cdev-states
  prefix: pref
```

### Secrets

There are two ways to use secrets:

#### SOPS secret

Secrets are encoded/decoded with [SOPS](https://github.com/mozilla/sops) utility that supports AWS KMS, GCP KMS, Azure Key Vault and PGP keys.
How to use:

1. Use console client cdev to create a new secret from scratch:

```bash
cdev secret create
```

2. Use interactive menu to create a secret.

3. Edit the secret and set secret data in `encrypted_data:` section.

4. Use references to the secret's data in infrastructure template (you can find the examples in the generated secret file).

#### Amazon secret manager

cdev client can use AWS SSM as a secret storage.

How to use:

1. create a new secret in AWS secret manager using AWS cli or web console. Both raw and JSON data formats are supported.

2. Use the console client cdev to create a new secret from scratch:

```bash
cdev secret create
```

3. Answer the questions. For `Name of secret in AWS Secrets manager` enter the name of the AWS secret created above.

4. Use references to the secret's data in infrastructure template (you can find the examples in the generated secret file).

To list and edit any secret, use the commands:

```bash
cdev secret ls
```

and

```bash
cdev secret edit secret_name
```

## Template configuration

### Basics

Template represents yaml structure with the array of different invocation [modules](#Modules)
Common view:

```yaml
modules:
  - module1
  - module2
  - module3
  ...
```

Template can utilize all kinds of go-templates and Sprig functions (similar to Helm). Along with that it is enhanced with functions like insertYAML that could pass yaml blocks directly.

### Functions

1) [Base go-template language functions](https://golang.org/pkg/text/template/#hdr-Functions).

2) [Sprig functions](https://masterminds.github.io/sprig/).

3) Enhanced functions: all functions described above allow you to modify the template text. Apart from these, some special enhanced functions are available. They cannot be used everywhere. The functions are integrated within the functionality of the program and with the yaml syntax:

* `insertYAML` - pass yaml block as value of target yaml template. **Argument**: data to pass, any value or reference to block.
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

* `remoteState` - is used for passing data between modules and infrastructures. **Argument**: string, path to remote state consisting of 3 parts separated by a dot: `"infra_name.module_name.output_name"`. Since the name of the infrastructure is unknown inside the template, you can use "this" instead:`"this.module_name.output_name"`. **Allowed use**: as yaml value, only in module `inputs`.

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

* `name` - module name. *Required*.

* `type` - module type. One of: `terraform`, `helm`, `kubernetes`, `printer`. See below.

* `depends_on` - *string* or *list of strings*. One or multiple module dependencies in the format "infra_name.module_name". Since the name of the infrastructure is unknown inside the template, you can use "this" instead:`"this.module_name.output_name"`.

* `pre_hook` and `post_hook` blocks: describe the shell commands to be executed before and after the module, respectively. The commands will be executed in the same context as the actions of the module. Environment variables are common to the shell commands, the pre_hook and post_hook scripts, and the module execution. You can export a variable in the pre_hook and it will be available in the post_hook or in the module.

  * `command` - *string*. Shell command in text format. Will be executed in bash -c "command". Can be used if the "script" option is not used. One of `command` or `script` is required.

  * `script` - *string* path to shell script file which is relative to template directory. Can be used if the "command" option is not used. One of `command` or `script` is required.

  * `on_apply` *bool*, *optional* turn off/on when module applying. **Default: "true"**.

  * `on_destroy` - *bool*, *optional* turn off/on when module destroying. **Default: "false"**.

  * `on_plan` - *bool*, *optional* turn off/on when module plan executing. **Default: "false"**.

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

* `source` - *string*, *required*. Terraform module [source](https://www.terraform.io/docs/language/modules/syntax.html#source). **It is not allowed to use local folders in source!**.

* `version` - *string*, *optional*. Module [version](https://www.terraform.io/docs/language/modules/syntax.html#version).

* `inputs` - *map of any*, *required*. A map that corresponds to [input variables](https://www.terraform.io/docs/language/values/variables.html) defined by the module. This block allows to use functions `remoteState` and `insertYAML`.

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

* `source` - *map*, *required*. Block describes helm chart source.

  * `chart`, `repository`, `version` - correspond to options with the same name from helm_release resource. See [chart](https://registry.terraform.io/providers/hashicorp/helm/latest/docs/resources/release#chart), [repository](https://registry.terraform.io/providers/hashicorp/helm/latest/docs/resources/release#repository) and [version](https://registry.terraform.io/providers/hashicorp/helm/latest/docs/resources/release#version).

  * `kubeconfig` - *string*, *required*. Path to the kubeconfig file which is relative to the directory where the module was executed.

  * `additional_options` - *map of any*, *optional*. Corresponds to [Helm_release resource options](https://registry.terraform.io/providers/hashicorp/helm/latest/docs/resources/release#argument-reference). Will be pass as is.

  * `inputs` - *map of any*, *optional*. A map that represents [Helm release sets](https://registry.terraform.io/providers/hashicorp/helm/latest/docs/resources/release#set). This block allows to use functions `remoteState` and `insertYAML`.
    For example:

    ```yaml
    inputs:
      global.image.tag: v1.8.3
      service.type: LoadBalancer
    ```

    corresponds to:

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

* `source` - *string*, *required*. Path to Kubernetes manifest that will be converted into a representation of kubernetes-alpha provider. **Source file will be rendered with the template, and also allows to use the functions `remoteState` and `insertYAML`**.

* `kubeconfig` - *string*, *required*. Path to the kubeconfig file which is relative to the directory where the module was executed.

#### Printer module

The module is mainly used to see the outputs of other modules in the console logs.

Example:

```yaml
modules:
  - name: print_outputs
    type: printer
    inputs:
      cluster_name: {{ .name }}
      worker_iam_role_arn: {{ remoteState "this.eks.worker_iam_role_arn" }}
```

* `inputs` - *any*, *required* - a map that represents data to be printed in the log. The block **allows to use the functions `remoteState` and `insertYAML`**.
