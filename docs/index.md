# Cluster.dev - Kubernetes-based Dev Environment in Minutes

Cluster.dev is an open-source system delivered as GitHub Action or Docker Image 
for creating and managing Kubernetes clusters with simple manifests by GitOps approach.   

Designed for developers that are bored to configure Kubernetes stuff
and just need: kubeconfig, dashboard, logging and monitoring out-of-the-box.  

Based on DevOps and SRE best-practices. GitOps cluster management and application delivery.
Simple CICD integration. Easily extandable by pre-configured applications and modules. 
Supports different Cloud Providers and Kubernetes versions.

----
## Principle diagram

![cluster.dev diagram](images/cluster-dev-diagram.png)


## How it works

In background:

 - Terraform creates a "state bucket" in your Cloud Provider account where all infrastructure objects would be stored. Typically it is defined on Cloud Object Storage like AWS S3.
 - Terraform modules creates Minikube/EKS/GKE/etc.. cluster, VPC and DNS zone within your Cloud Provider.
 - ArgoCD Continuous Deployment system deployed inside Kubernetes cluster. Enables you to deploy your applications from raw manifests, helm charts or kustomize yaml's.
 - GitHub CI runner deployed into your Kubernetes cluster and used for your apps building CI pipelines with GitHub Actions.

You receive:

 - Automatically generated kubeconfig, ssh-access, and ArgoCD UI url’s
 - Configured: Ingress Load Balancers, Kubernetes Dashboard, Logging(ELK), Monitoring(Prometheus/Grafana)  

## Quick Start

Sample manifest to create cluster:
```yaml
cluster:
  name: minikube-a
  cloud: 
    provider: aws
    region: eu-central-1
    vpc: default
    domain: shalb.net
  provisioner:
    type: minikube
    instanceType: m5.large
```   
You can find complete sample in our [GitHub Repo/Quick Start](https://github.com/shalb/cluster.dev#quick-start)

## Roadmap 

The project is in Alpha Stage, the first release is planned on 14 February 2020.
Roadmap details: [ROADMAP](./roadmap/)

## Contributing 

If you want to spread the project with your own code, you can start contributing with this quick guide: [CONTRIBUTING](./contributing/)
