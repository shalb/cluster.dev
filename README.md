# Create Complete Kubernetes dev Enviornment
## Cluster.dev 
Container Image and GitHub Action to create and manage Kubernetes clusters with simple manifests.

Designed for developers that are bored to configure all that Kubernetes stuff and just need kubeconfig, dashboard, logging and monitoring out of the box.

Based on best-practices with GitOps application devivery. 

Easy extandable by pre-condigured applications and modules. Supports different Cloud Providers.

## Quick Start

Just create file in your repository  `.cluster.dev/minikube-a.yaml` 
```yaml
cluster:
  name: minikube-a
  cloud: 
    provider: aws
    region: eu-central-1
  provisioner:
    type: minikube
    instanceType: "m4.large"
```


Add a GitHub Workflow: `.github/workflows/main.yml`:  
```yaml
on: [push]
jobs:
  deploy_cluster_job:
    runs-on: ubuntu-latest
    name: Deploy and Update K8s Cluster
    steps:
    - name: Checkout Repo
      uses: actions/checkout@v1
    - name: Reconcile Clusters
      id: reconcile
      uses: shalb/cluster.dev@master
      with:
        cluster-config: './.cluster.dev/minikube-one.yaml'
        cloud-user: ${{ secrets.aws_access_key_id }}
        cloud-pass: ${{ secrets.aws_secret_access_key }}
    - name: Get the execution status
      run: echo "The status ${{ steps.validate.reconcile.status }}"
```

Also you need to add cloud credentials to your repo secrets, ex: 
```yaml
aws_access_key_id =  ATIAAJSXDBUVOQ4JR
aws_secret_access_key = SuperAwsSecret
```

Thats it! Just push update and cluster.dev would create you a cluster in minutes.

## How it works

In background: 
 - Terraform create remote state
 - Terraform modules creates Minikube instance with AWS Account credentials
 - Produced kubeconfig should be generated and passed to value into target git repo credentials
