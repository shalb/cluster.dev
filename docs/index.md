# What is Cluster.dev?

Cluster.dev is an open-source tool designed to manage cloud native infrastructures with simple declarative manifests - stack templates. It allows you to describe an entire infrastructure and deploy it with a single tool.

[Stack templates](https://docs.cluster.dev/stack-templates-overview/) can be based on Terraform modules, Kubernetes manifests, Shell scripts, Helm charts and Argo CD/Flux applications, OPA policies, etc. Cluster.dev brings those components together so that you can deploy, test and distribute a whole set of components with pinned versions.

## When do I need Cluster.dev?

1. If you have a common infrastructure pattern that contains multiple components stuck together. This could be a bunch of TF-modules, or a set of K8s add-ons where you need to re-use this pattern inside your projects.
2. If you develop an infrastructure platform that you ship to other teams and they need to launch new infrastructures from your template.
3. If you build a complex infrastructure that contains different technologies and you need to perform integration testing to confirm the components' interoperability. Once done, you can then promote the changes to next environments.
4. If you are a software vendor and need to deliver infrastructure deployment along with your software.

## Base concept diagrams

Stack templates are composed of [units](https://docs.cluster.dev/units-overview/) - Lego-like building blocks responsible for passing variables to a particular technology.

![cdev unit example diagram](./images/cdev-unit-example.png)

Templates define infrastructure patterns or even the whole platform.

![cdev template example diagram](./images/cdev-template-example.png)

## Features

- Common variables, secrets, and templating for different technologies.
- Same GitOps Development experience for Terraform, Shell, Kubernetes.
- Can be used with any cloud, on-premises or hybrid scenarios.
- Encourage teams to follow technology best practices.
