# Cluster.dev - Kubernetes-based Dev Environment in Minutes

Cluster.dev is an open-source system delivered as GitHub Action or Docker Image for creating and managing Kubernetes clusters with simple manifests by GitOps approach.

Designed for developers that are bored to configure Kubernetes stuff and just need: kubeconfig, dashboard, logging and monitoring out-of-the-box.

Based on DevOps and SRE best-practices. GitOps cluster management and application delivery. Simple CI/CD integration. Easily extendable by pre-configured applications and modules. Supports different Cloud Providers and Kubernetes versions.

----

## Principle diagram

![cluster.dev diagram](docs/images/cluster-dev-diagram.png)


## How it works

In background:

- Terraform creates a "state bucket" in your Cloud Provider account where all infrastructure objects will be stored. Typically it is defined on Cloud Object Storage like AWS S3.
- Terraform modules create Minikube/EKS/GKE/etc.. cluster, VPC and DNS zone within your Cloud Provider.
- ArgoCD Continuous Deployment system is deployed inside Kubernetes cluster. It enables you to deploy your applications from raw manifests, helm charts or kustomize yaml's.
- GitHub CI runner is deployed into your Kubernetes cluster and used for your apps building CI pipelines with GitHub Actions.

You receive:

- Automatically generated kubeconfig, ssh-access, and ArgoCD UI URLs
- Configured: Ingress Load Balancers, Kubernetes Dashboard, Logging(ELK), Monitoring(Prometheus/Grafana)

## Quick Start

1. Dedicate a separate repository for the infrastructure that will be managed by `cluster.dev`. This repo will host code for your clusters, deployments and other resources managed by GitOps.  
Next steps should be done in that repo.

2. Obtain access credentials for your cloud account.  
For example, in AWS it is called "Programmatic Access user", and looks like:

```yaml
aws_access_key_id = ATIAAJSXDBUVOQ4JR
aws_secret_access_key = SuperAwsSecret
```

3. Add credentials to you repo Secrets under GitHub's: "Settings->Secrets", ex:
 ![GitHub Secrets](docs/images/gh-secrets.png)

4. Create a new cluster.dev yaml manifest with your cluster definition, example: `.cluster.dev/minikube-one.yaml`:

```yaml
cluster:
  installed: true
  name: minikube-five
  cloud:
    provider: aws
    region: eu-central-1
    vpc: default
    domain: shalb.net  # You need a domain in Route53
  provisioner:
    type: minikube
    instanceType: m5.large
  # 'apps' are actually directories in '//kubernetes/apps'
  # that contains ArgoCD project and application declarations
  apps:
    - helm-all-in-app
    - helm-dependency
    - raw-manifest
```

5. Create a Github Workflow file `.github/workflows/main.yml`:

```yaml
on:
  push:
    paths:
      - '.github/workflows/**' # path to workflow
      - '.cluster.dev/**' # Path to cluster declaration manifests
      - '//kubernetes/apps/**' # Create your ArgoCD app folders here
    branches:
      - master
jobs:
  deploy_cluster_job:
    runs-on: ubuntu-latest
    name: Create and Update Kubernetes Clusters
    steps:
    # To use this repository's private action, you must check out the repository
    - name: Checkout Repo
      uses: actions/checkout@v1
    # - uses: ./ # Uses an action in the root directory
    - name: Reconcile Clusters
      id: reconcile
      uses: shalb/cluster.dev@master
      with:
        cluster-config: './.cluster.dev/minikube-one.yaml'
        cloud-user: ${{ secrets.aws_access_key_id }}
        cloud-pass: ${{ secrets.aws_secret_access_key }}
    # Use the output from the `reconcile` step
    - name: Get the Cluster Credentials
      run: echo -e "\n\033[1;32m${{ steps.reconcile.outputs.ssh }}\n\033[1;32m${{ steps.reconcile.outputs.kubeconfig }}\n\033[1;32m${{ steps.reconcile.outputs.argocd }}"
```

6. Commit and Push files to your repo and follow the Github Action execution status. In GitHub action output you'll receive access instructions to your cluster and services.

## Technical diagram

![cluster.dev technical diagram](docs/images/cluster-dev-technical-diagram.png)


## Roadmap

The project is in Alpha Stage. Roadmap details: [docs/ROADMAP.md](docs/ROADMAP.md)

## Contributing

If you want to spread the project with your own code, you could start contributing with this quick guide: [docs/CONTRIBUTING.md](docs/CONTRIBUTING.md)
