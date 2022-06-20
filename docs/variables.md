# Variables

Stack configuration contains variables that will be applied to a stack template (similar to `values.yaml` in Helm or `tfvars` file in Terraform). The variables from `stack.yaml` are passed to stack template files to render them.  

Example of `stack.yaml` with variables `region`, `name` `organization`:

```yaml
name: k3s-demo
template: https://github.com/shalb/cdev-s3-web
kind: Stack
variables:
  region: eu-central-1
  name: web-static-page
  organization: Cluster.dev
```

The values of the variables are passed to a `template.yaml` to configure the resulting infrastructure.  

## Variables reference

Variables could be used in all configurations of stacks and backends within a given project.

To refer to a variable, use the {{ .project.variables.KEY_NAME }} syntax, where `.project.variables` is the path that follows the structure of variables in a `project.yaml`:

```yaml
name: demo
kind: Project
variables:
  region: eu-central-1
```

The KEY_NAME stands for the variable name defined in the `project.yaml` and will be replaced by its value. Example of variables reference in a `stack.yaml`:

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

## Passing variables across stacks and units

Cluster.dev allows passing variables across different stacks, and across units within one stack template. This is made possible in 2 ways:

* using the output of one unit/stack as an input for another unit/stack: {{ output "stack_name.unit_name.output" }}

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

!!! note
      If passing outputs across units within one stack template, use "this" instead of the stack name: {{ output "this.unit_name.output" }}:.

* using [`remoteState`](https://docs.cluster.dev/stack-templates-functions/#remotestate) with a syntax: {{ remoteState "stack_name.unit_name.output" }}

  Example of passing variables across units in the stack template:

  ```yaml
  name: s3-static-web
  kind: StackTemplate
  units:
    - name: s3-web
      type: tfmodule
      source: "terraform-aws-modules/s3-bucket/aws"
      providers:
      - aws:
          region: {{ .variables.region }}
      inputs:
        bucket: {{ .variables.name }}
        force_destroy: true
        acl: "public-read"
    - name: outputs
      type: printer
      outputs:
        bucket_name: {{ remoteState "this.s3-web.s3_bucket_website_endpoint" }}
        name: {{ .variables.name }}
  ```

## Global and stack-specific variables

The variables stored and defined within a project are global. They could be common for a few stacks within one project and [passed across these stacks](#passing-variables-across-units-and-stacks). Example of global variables in the `project.yaml`:

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

Global variables can’t be used in stack templates directly. From `project.yaml` the variable value is passed to `stack.yaml` from where it is applied to a stack template. See [Templating](https://docs.cluster.dev/templating) for more details.  

Stack-specific variables are defined within a stack and relate to a concrete infrastructure. They can be used solely in the stack templates that are bound to this stack.


