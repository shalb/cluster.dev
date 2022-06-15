# Variables

## Global and stack-specific variables

The variables stored and defined within a project are global. They could be common for a few stacks within one project and [passed across these stacks](#passing-variables-across-stacks). Example of global variables in the `project.yaml`:

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

## Variables reference

Global variables could be used in all configurations of stacks and backends within a given project.

To refer to a global variable, use the {{ .project.variables.KEY_NAME }} syntax, where `.project.variables` is the path that follows the structure of variables in a `project.yaml`:

```yaml
name: demo
kind: Project
variables:
  region: eu-central-1
```

The KEY_NAME stands for the variable name defined in the `project.yaml` and will be replaced by its value. Example of global variables reference in a `stack.yaml`:

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

The same syntax applies to the secrets' data: {{ .secrets.secret_name.secret_key }}. Let’s assume we have a secret in AWS Secrets Manager:

```yaml
name: my-aws-secret
kind: Secret
driver: aws_secretmanager
spec: 
    region: eu-central-1
    aws_secret_name: pass
```

In order to refer to the secret in stack files, we need to define it as {{ .secrets.my-aws-secret.some-key }}.

## Passing variables across units and stacks

Cluster.dev allows to pass variables both across units within one stack template, and between different stacks. Both options are implemented by the same mechanisms:

* using [`remoteState` function](https://docs.cluster.dev/stack-templates-functions/#remotestate) with a syntax: {{ remoteState "stack_name.unit_name.output" }}

* using one unit/stack output as an input for another unit/stack: {{ output "stack_name.unit_name.output" }}