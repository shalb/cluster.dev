# Cluster.dev - Working Principles

With Cluster.dev you download a predefined stack template, set the variables, then render and deploy a whole stack.

Capabilities:

- Re-using all existing Terraform private and public modules and Helm Charts.
- Applying parallel changes in multiple infrastructures concurrently.
- Using the same global variables and secrets across different infrastructures, clouds and technologies.
- Templating anything with Go-template function, even Terraform modules in Helm style templates.
- Create and manage secrets with SOPS or cloud secret storages.
- Generate a ready-to-use Terraform code.

## Basic diagram

![cdev diagram](./images/cdev-base-diagram.png)

## Variables

Cluster.dev uses global and stack-specific variables.

Global variables are defined within project. They could be common for a few stacks that are reconciled within a project, and passed across them. Example of `project.yaml`:

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

 From `project.yaml` the variable value is passed to `stack.yaml` from where it is applied to a stack template.
 
 Global variables could be used in all configurations of stacks and backends within a given project. To refer to a global variable, use the {{ .project.variables.KEY_NAME }} syntax, where KEY_NAME stands for the variable name defined in `project.yaml` and will be replaced by its value. Example of global variables in `stack.yaml`:

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

Stack-specific variables are defined in `stack.yaml` and relate to a concrete infrastructure. They can be used solely in stack templates that are bound to this stack.

## How to use Cluster.dev

Cluster.dev is quite a powerful framework that can be operated in several modes.

### Deploy infrastructures from existing stack templates

This mode, also known as **user mode**, gives you the ability to launch ready-to-use infrastructures from prepared stack templates by just adding your cloud credentials and setting variables (such as name, zones, number of instances, etc.).
You don't need to know background tooling like Terraform or Helm, it's just as simple as downloading a sample and launching commands. Here are the steps:

* Install Cluster.dev binary
* Choose and download a stack template
* Set cloud credentials
* Define variables for the stack template
* Run Cluster.dev and get a cloud infrastructure

### Create your own stack template

In this mode you can create your own stack templates. Having your own template enables you to launch or copy environments (like dev/stage/prod) with the same template.
You'll be able to develop and propagate changes together with your team members, just using Git.
Operating Cluster.dev in the **developer mode** requires some prerequisites. The most important is understanding Terraform and how to work with its modules. The knowledge of `go-template` syntax or `Helm` is advisable but not mandatory.

The easiest way to start is to download/clone a sample template project like [AWS-EKS](https://github.com/shalb/cdev-aws-eks)
and launch an infrastructure from one of the examples.
Then you can edit some required variables, and play around by changing values in the template itself.

#### Workflow

Let's assume you are starting a new infrastructure project. Let's see how your workflow would look like.

1. Define what kind of infrastructure pattern you need to achieve.

      a. What Terraform modules it would include (for example: I need to have VPC, Subnet definitions, IAM's and Roles).
    
      b. Whether you need to apply any Bash scripts before and after the module, or inside as pre/post-hooks.
    
      c. If you are using Kubernetes, check what controllers would be deployed and how (by Helm chart or K8s manifests).

2. Check if there is any similar sample template that already exists.

3. Clone the stack template locally.

4. Apply it.

