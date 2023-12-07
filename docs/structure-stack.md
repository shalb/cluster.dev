# Stack

Stack is a yaml file that tells Cluster.dev which template to use and what [variables](https://docs.cluster.dev/templating/#variables) to apply to this template. Usually, users have multiple stacks that reflect their environments or tenants, and point to the same template with different variables.

File: searching in `./*.yaml`. *Required at least one*.
Stack object (`kind: stack`) contains reference to a stack template, variables to render the template and backend for states.

Example of `stack.yaml`:

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

* `name`- *Required*. The stack name. 

* `kind`- *Required*. Object kind `stack`. 

* `backend`- *Optional*. Name of the backend that will be used to store the states of this stack. 

* `variables`- data set for the stack template rendering. See [variables](https://docs.cluster.dev/templating/#variables).

*  `template`- *Required*. Either a path to a local directory containing the stack template's configuration files, or a remote Git repository as the stack template source. For more details on stack templates please refer to [Stack Template](https://docs.cluster.dev/stack-templates-overview/) section. A local path must begin with either `/` for absolute path, `./` or `../` for relative path. For Git source, use this format: `<GIT_URL>//<PATH_TO_TEMPLATE_DIR>?ref=<BRANCH_OR_TAG>`:
    * `<GIT_URL>` - *required*. Standard Git repo url. See details on [official Git page](https://git-scm.com/docs/git-clone#_git_urls).
    * `<PATH_TO_TEMPLATE_DIR>` - *optional*, use it if the stack template's configuration is not in repo root.
    * `<BRANCH_OR_TAG>`- Git branch or tag.

* `disabled`- *bool*, *optional*. Disable stack execution. By default is set to `false`. If set to `true` the stack won't be applied. 

## Examples

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
