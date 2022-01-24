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

## Templating

Templating is one of the key features that underlie powerful capabilities of Cluster.dev. Same like with Helm, the cdev templating is based on Go template language and uses Sprig and some other extra functions to expose objects to the templates.

Cluster.dev has a two-level templating that involves template rendering on a project level and on a stack template level. For more information please refer to the [Templating section](https://docs.cluster.dev/templating/).

## How to use Cluster.dev

Cluster.dev is a powerful framework that can be operated in several modes.

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

