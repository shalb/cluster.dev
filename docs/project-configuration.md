# Project Configuration

Common project files:

```bash
project.yaml        # Contains global project variables that can be used in other configuration objects.
<infra_name>.yaml   # Contains reference to a stack template, variables to render the stack template and backend for states.
<backend_name>.yaml # Describes a backend storage for Terraform and cdev states.
<secret_name>.yaml  # Contains secrets, one per file.
```

Cdev reads configuration from current directory, i.e. all files by mask: `*.yaml`. It is allowed to place several yaml configuration objects in one file, separating them with "---". The exception is the project.yaml configuration file and files with secrets.

## Project

Project is a storage for global variables related to all stacks. It is a high-level abstraction to store and reconcile different stacks, and pass values across them.

File: `project.yaml`. *Required*.
Represents a set of configuration options for the whole project. Contains global project variables that can be used in other configuration objects, such as backend or stack (except of `secrets`). Note that the `project.conf` file is not rendered with the template and you cannot use template units in it.

Example `project.yaml`:

```yaml
name: my_project
kind: project
backend: aws-backend
variables:
  organization: shalb
  region: eu-central-1
  state_bucket_name: cdev-states
exports:
  AWS_PROFILE: cluster-dev  
```

* `name`: project name. *Required*.

* `kind`: object kind. Must be set as `project`. *Required*.

* `backend`: name of the backend that will be used to store the cdev state of the current project. *Optional*. If the backend is not specified the state will be saved locally in the ./cdev.state file. For now only S3 bucket backends are supported. 

* `variables`: a set of data in yaml format that can be referenced in other configuration objects. For the example above, the link to the organization name will look like this: `{{ .project.variables.organization }}`.

* `exports`: list of environment variables that will be exported while working with the project. *Optional*.

## Stack

Stack is a yaml file that tells cdev which template to use and what variables to apply to this template. Usually, users have multiple stacks that reflect their environments or tenants, and point to the same template with different variables.

File: searching in `./*.yaml`. *Required at least one*.
Stack object (`kind: stack`) contains reference to a stack template, variables to render the template and backend for states.

Example:

```yaml
# Define stack itself
name: k3s-infra
template: "./templates/"
kind: stack
backend: aws-backend
variables:
  bucket: {{ .project.variables.state_bucket_name }} # Using project variables.
  region: {{ .project.variables.region }}
  organization: {{ .project.variables.organization }}
  domain: cluster.dev
  instance_type: "t3.medium"
  vpc_id: "vpc-5ecf1234"
```

* `name`: stack name. *Required*.

* `kind`: object kind. `stack`. *Required*.

* `backend`: name of the backend that will be used to store the states of this stack. *Required*.

* `variables`: data set for a stack template rendering.

*  <a name="infra_options_template">`template`</a>: it's either a path to a local directory containing the stack template's configuration files, or a remote Git repository as the stack template source. For more details on stack templates please see the [Stack Template Development](https://cluster.dev/stack-template-development/) section. A local path must begin with either `/` for absolute path, `./` or `../` for relative path. For Git source, use this format: `<GIT_URL>//<PATH_TO_TEMPLATE_DIR>?ref=<BRANCH_OR_TAG>`:
    * `<GIT_URL>` - *required*. Standard Git repo url. See details on [official Git page](https://git-scm.com/docs/git-clone#_git_urls).
    * `<PATH_TO_TEMPLATE_DIR>` - *optional*, use it if the stack template's configuration is not in repo root.
    * `<BRANCH_OR_TAG>`- Git branch or tag.

Examples:

```yaml
template: /path/to/dir # absolute local path
template: ./template/ # relative local path
template: ../../template/ # relative local path
template: https://github.com/shalb/cdev-k8s # https Git url
template: https://github.com/shalb/cdev-k8s//some/dir/ # subdirectory
template: https://github.com/shalb/cdev-k8s//some/dir/?ref=branch-name # branch
template: https://github.com/shalb/cdev-k8s?ref=v1.1.1 # tag
template: git@github.com:shalb/cdev-k8s.git # ssh Git url
template: git@github.com:shalb/cdev-k8s.git//some/dir/ # subdirectory
template: git@github.com:shalb/cdev-k8s.git//some/dir/?ref=branch-name # branch
template: git@github.com:shalb/cdev-k8s.git?ref=v1.1.1 # tag
```

## Backends

Backend is an object that describes backend storage for Terraform and cdev states.

File: searching in `./*.yaml`. *Required at least one*.
In the backends' configuration you can use any options of the appropriate Terraform backend. They will be converted as is.
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

* `azurerm` Microsoft azurerm:

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

## Secrets

Secret is an object that contains sensitive data such as a password, a token, or a key. Is used to pass secret values to the tools that don't have a proper support of secret engines.

There are two ways to use secrets:

### SOPS secrets

For **creating** and **editing** SOPS secrets, cdev uses SOPS binary. But the SOPS binary is **not required** for decrypting and using SOPS secrets. As none of cdev reconcilation processes (build, plan, apply) requires SOPS to be performed, you don't have to install it for pipelines.

See [SOPS installation instructions](https://github.com/mozilla/sops#download) in official repo.

Secrets are encoded/decoded with [SOPS](https://github.com/mozilla/sops) utility that supports AWS KMS, GCP KMS, Azure Key Vault and PGP keys. How to use:

1. Use console client cdev to create a new secret from scratch:

     ```bash
     cdev secret create
     ```

2. Use interactive menu to create a secret.

3. Edit the secret and set secret data in `encrypted_data:` section.

4. Use references to the secret's data in a stack template (you can find the examples in the generated secret file).

### Amazon secret manager

Cdev client can use AWS SSM as a secret storage. How to use:

1. Create a new secret in AWS secret manager using AWS CLI or web console. Both raw and JSON data formats are supported.

2. Use the console client cdev to create a new secret from scratch:

     ```bash
     cdev secret create
     ```

3. Answer the questions. For `Name of secret in AWS Secrets manager` enter the name of the AWS secret created above.

4. Use references to the secret's data in a stack template (you can find the examples in the generated secret file).

To list and edit any secret, use the commands:

```bash
cdev secret ls
```

and

```bash
cdev secret edit secret_name
```

## Stack templates

Currently there are 3 types of stack templates available:

  * [aws-k3s](https://github.com/shalb/cdev-aws-k3s)
  * [aws-eks](https://github.com/shalb/cdev-aws-eks)
  * [do-k8s](https://github.com/shalb/cdev-do-k8s)

For the detailed information on templates, please see the section [Stack Template Development](https://cluster.dev/stack-template-development/).
