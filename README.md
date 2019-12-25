# Create Complete Kubernetes dev Enviornment
## Deploy Cluster

Create Kubernetes Cluster Based on manifest:
```yaml
cluster:
  name: minikube-one
  cloud: 
    provider: aws
    region: eu-central-1
  provisioner:
    type: minikube
    instanceType: "m4.large"
```
Cluster should be deployed by adding `.cluster.dev/minikube-one.yaml` into repository
And also requires a GitHub Workflow to be added to `.github/workflows/main.yml`:

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

In background there should be: 
 - Terraform module which creates Minikube instance with AWS Account credentials
 - Produced kubeconfig should be generated and passed to value into target git repo credentials
