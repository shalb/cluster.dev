# Cluster.dev - Cloud infrastructures' management tool

## What is it?

Cluster.dev is an open-source tool designed to manage Cloud Native Infrastructures with simple declarative manifests - infrastructure templates. It allows you to describe a whole infrastructure and deploy it with a single tool. 

The infrastructure templates could be based on Terraform modules, Kubernetes manifests, Shell scripts, Helm charts, Kustomize and ArgoCD/Flux applications, OPA policies etc.. Cluster.dev sticks those components together so that you could deploy, test and distribute a whole set of components with pinned versions. 

## Principle diagram

![cluster.dev diagram](images/cluster-dev-diagram.png)

## How it works

In the background:

- Infrastructures are described as simple [infrastructure manifests](https://github.com/shalb/cluster.dev/tree/master/.cluster.dev) and are stored in a Git repository.
- Infrastructure changes are watched by GitHub/GitLab/Bitbucket pipeline and trigger the launch of the reconciler tool.
- Reconciler tool generates Terraform variables files and performs ordered invoking for the modules.
- Terraform creates a "state bucket" in your Cloud Provider account where all infrastructure objects and configs are stored. Typically it is defined on Cloud Object Storage like AWS S3.
- Terraform modules create Minikube/EKS/GKE/etc.. cluster, VPC and DNS zone within your Cloud Provider.
- Kubernetes addons module deploys: Ingress controller, Cert-Manager, External DNS, ArgoCD, Keycloak, etc..
- ArgoCD continuous deployment system watches remote Git repositories and deploys your applications from raw manifests, Helm charts or Kustomize yamls.

You receive:

- Automatically generated kubeconfig, ArgoCD UI URL's.
- Pre-configured: VPC, Networks, Domains, Security groups, Users, etc..
- Deployed inside Kubernetes: Ingress Load Balancers, Kubernetes Dashboard, Logging (ELK), Monitoring (Prometheus/Grafana).

## Features

- Based on DevOps and SRE best-practices.
- Simple CI/CD integration.
- GitOps cluster management and application delivery.
- Automated provisioning of Kubernetes clusters in AWS, DO and GCE(in progress).

## Roadmap

The cluster.dev project is in Alpha Stage. You can check its progress and upcoming features on the [roadmap page](ROADMAP.md).
