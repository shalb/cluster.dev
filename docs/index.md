# Cluster.dev - Kubernetes-based environments in minutes

## What is it?

Cluster.dev is the global cloud orchestration service for modern infrastructures. It is an open-source system delivered as GitHub Action or Docker Image
for creating and managing Kubernetes clusters with simple manifests by GitOps approach.

Deploy fully managed Kubernetes clusters across AWS, Azure, or GCP with a few clicks. Best-in-class automation and proven practices guarantee availability, scalability, and compliance with the most demanding data security and privacy standards.

Designed for developers who are bored to configure Kubernetes stuff and just need: kubeconfig, dashboard, logging and monitoring out-of-the-box.

## How it works

In the background:

- Terraform creates a "state bucket" in your Cloud Provider account where all infrastructure objects will be stored. Typically it is defined on Cloud Object Storage like AWS S3.
- Terraform modules create Minikube/EKS/GKE/etc.. cluster, VPC and DNS zone within your Cloud Provider.
- ArgoCD Continuous Deployment system is deployed inside Kubernetes cluster. It enables you to deploy your applications from raw manifests, helm charts or kustomize yaml's.
- GitHub CI runner is deployed into your Kubernetes cluster and is used for building CI pipelines for your apps with GitHub Actions.

You receive:

- Automatically generated kubeconfig, SSH-access, and ArgoCD UI URLs.
- Configured: Ingress Load Balancers, Kubernetes Dashboard, Logging (ELK), Monitoring (Prometheus/Grafana).

## Principle diagram

![cluster.dev diagram](images/cluster-dev-diagram.png)

## Features

- Based on DevOps and SRE best-practices.
- Simple CI/CD integration.
- GitOps cluster management and application delivery.
- Automated provisioning of Kubernetes clusters in AWS, DO and GCE (in progress).

## Roadmap

The cluster.dev project is in Alpha Stage. You can check its progress and upcoming features on the [roadmap page](ROADMAP.md).
