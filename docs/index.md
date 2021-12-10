# Cluster.dev - Cloud Infrastructures' Management Tool

Cluster.dev is an open-source tool designed to manage cloud native infrastructures with simple declarative manifests - stack templates. It allows you to describe a whole infrastructure and deploy it with a single tool.

[Stack templates](https://docs.cluster.dev/stack-templates-overview/) could be based on Terraform modules, Kubernetes manifests, Shell scripts, Helm charts and ArgoCD/Flux applications, OPA policies, etc. Cluster.dev sticks those components together so that you could deploy, test and distribute a whole set of components with pinned versions.

## When do I need Cluster.dev?

1. If you have a common infrastructure pattern that contains multiple components stuck together.
   Like a bunch of TF-modules, or a set of K8s addons. So you need to re-use this pattern inside your projects.
2. If you develop an infrastructure platform that you ship to other teams, and they need to launch new infras from your template.
3. If you build a complex infrastructure that contains different technologies, and you need to perform integration testing to confirm the components' interoperability. After which you can promote the changes to next environments.
4. If you are a software vendor and you need to deliver infrastructure deployment along with your software.

## Base concept diagrams

Stack templates are composed of [units](https://docs.cluster.dev/units-overview/) - Lego-like building blocks responsible for passing variables to a particular technology.

![cdev unit example diagram](./images/cdev-unit-example.png)

Templates define infrastructure patterns or even the whole platform.

![cdev template example diagram](./images/cdev-template-example.png)

## Features

- Common variables, secrets and templating for different technologies.
- Same GitOps Development experience for Terraform, Shell, Kubernetes.
- Could be used with any Cloud, On-premises or Hybrid scenarios.
- Encourage teams to follow technology best practices.
