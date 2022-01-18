# Templating

## Levels of templating

Cluster.dev has a two-level templating that is applied on the project's and on the stack template's levels. 

On the first level Cluster.dev reads a `project.yaml` and files with secrets. Then it uses the variables from these files to populate and render files from the current project – stacks and backends. These variables are global and could be passed across stacks and backends within a single project.

As global variables can’t be used in stack templates directly, the stack template files are rendered from data contained in the stack object (an outcome of the first stage). This is the stack level templating. 

The templating process could be described as follows: 

1.	Reading data from `project.yaml` and secrets. 

2.	Using the data to render all `yaml` files within the project directory. 

3.	Reading data from `stack.yaml` and `backend.yaml` (the files rendered in p.#2) – **first-level templating**.

4.	Downloading specified stack templates. 

5.	Rendering the stack templates from data contained in the corresponding `stack.yaml` files (p.#3) – **second-level templating**.

6.	Reading units from the stack templates.

7.	Executing the project.  

## Variables reference

To use data from a `project.yaml` in templating, define the path as `.project.variables`. Such a path follows the structure of variables in project.yaml, for example:

```yaml
name: demo
kind: Project
variables:
  region: eu-central-1
```
We can refer to `region` variable in stack files by defining it as {{ .project.variables.region }}.

The same applies to the secrets' data: {{ .secrets.secret_name.secret_key }}. Let’s assume we have a secret in AWS Secrets Manager:

```yaml
name: my-aws-secret
kind: Secret
driver: aws_secretmanager
spec: 
    region: eu-central-1
    aws_secret_name: pass
```

In order to refer to the secret in stack files, we need to define it as {{ .secrets.my-aws-secret.some-key }}.
