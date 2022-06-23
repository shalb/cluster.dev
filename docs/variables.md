# Variables

Stack configuration contains variables that will be applied to a stack template (similar to `values.yaml` in Helm or `tfvars` file in Terraform). The variables from `stack.yaml` are passed to stack template files to render them.  

Example of `stack.yaml` with variables `region`, `name`, `organization`:

```yaml
name: k3s-demo
template: https://github.com/shalb/cdev-s3-web
kind: Stack
variables:
  region: eu-central-1
  name: web-static-page
  organization: Cluster.dev
```

The values of the variables are passed to a stack template to configure the resulting infrastructure.  

## Passing variables across stacks 

Cluster.dev allows passing variable values across different stacks within one project. This is made possible in 2 ways:

* using the output of one stack as an input for another stack: {{ output "stack_name.unit_name.output" }}

  Example of passing outputs between stacks:

  ```yaml
  name: s3-web-page
  template: ../web-page/
  kind: Stack
  variables:
    region: eu-central-1
    name: web-static-page
    organization: Shalb
  ```

  ```yaml
  name: health-check
  template: ../health-check/
  kind: Stack
  variables:
    url: {{ output "s3-web-page.outputs.url" }}
  ```

* using [`remoteState`](https://docs.cluster.dev/stack-templates-functions/#remotestate) with a syntax: {{ remoteState "stack_name.unit_name.output" }}

## Global variables

The variables defined on a project level are called global. Global variables are listed in the `project.yaml`– a configuration file that defines the parameters and settings for the whole project. From the `project.yaml` the variables values can be applied to all stack and backend objects within this project. 

Example of the `project.yaml` file that contains variables `organization` and `region`:

```yaml
name: demo
kind: Project
variables:
  organization: shalb
  region: eu-central-1
```

To refer to a variable in stack and backend files, use the {{ .project.variables.KEY_NAME }} syntax, where *project.variables* is the path that corresponds the structure of variables in the `project.yaml`. The KEY_NAME stands for the variable name defined in the `project.yaml` and will be replaced by its value. 

Example of the `stack.yaml` file that contains reference to the project variables `organization` and `region`:

```yaml
name: eks-demo
template: https://github.com/shalb/cdev-aws-eks?ref=v0.2.0
kind: Stack
backend: aws-backend
variables:
  region: {{ .project.variables.region }}
  organization: {{ .project.variables.organization }}
  domain: cluster.dev
  instance_type: "t3.medium"
  eks_version: "1.20"
```

Example of the rendered `stack.yaml`:

```yaml
name: eks-demo
template: https://github.com/shalb/cdev-aws-eks?ref=v0.2.0
kind: Stack
backend: aws-backend
variables:
  region: eu-central-1
  organization: shalb
  domain: cluster.dev
  instance_type: "t3.medium"
  eks_version: "1.20"
```


