# Cluster.dev - Kubernetes-based Dev Environment in Minutes

Cluster.dev is open source system delivered as GitHub Action or Docker Image 
for creating and managing Kubernetes clusters with simple manifests by GitOps approach.  
Designed for developers that are bored to configure Kubernetes stuff
and just need: kubeconfig, dashboard, logging and monitoring out of the box.  

Based on DevOps and SRE best-practices. GitOps cluster management and application delivery.
Simple CICD integration. Easy extandable by pre-condigured applications and modules. 
Supports different Cloud Providers and Kubernetes versions.

----
## Principle diagram. What it does?

![cluster.dev diagram](docs/images/cluster-dev-diagram.png)

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

Thats it! Just push update and Cluster.dev would create you a cluster in minutes.
And produce a working kubeconfig that could be downloaded and links to differnet UI's: Kibana, Grafana, Dashboard, etc...

## How it works

In background: 

 - Terraform creates a remote state file where all infrastructure objects are stored.
   Typically it is defined on Cloud Object Storage like AWS S3.
 - Terraform modules creates Minikube/EKS/GKE/etc.. cluster within your Cloud Proivder using      Account credentials
 - Produced kubeconfig should be generated and passed to value into target git repo credentials

## Contributing 

If you want to spread the project with own code, you could start contribute following next instructions: [docs/CONTRIBUTING.md](docs/CONTRIBUTING.md)
