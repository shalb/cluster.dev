# Cluster.dev - Kubernetes infrastructures in minutes!

## What is it?

Cluster.dev is the cloud-native infrastructure orchestration framework. It is an open-source system delivered as a runtime inside Docker container.  

It is used for creating and managing Kubernetes clusters along with cloud resources like networks, domains and users with pre-defined Terraform modules.
The orchestration is performed with simple manifests by GitOps approach and is designed to run inside the GitHub/GitLab/BitBucket pipelines.  

Resulting infrastructures have a "ready to use" Continuous Deployment systems that could deploy manifests, Helm charts and Kustomize using ArgoCD.  

Best-in-class automation and proven practices guarantee availability, scalability, and compliance with the most demanding data security and privacy standards.
Designed for developers who are bored to configure cloud-native stack and just need infrastructure in code, kubeconfig, CD, dashboard, logging and monitoring out-of-the-box.

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
