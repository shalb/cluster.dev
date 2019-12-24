# Create Complete Kubernetes dev Enviornment
## Deploy Cluster

Create Kubernetes Cluster Based on manifest:
```yaml
cluster:
  name: minikube-one
  provider:
    minikube:             
      aws:
        instanceType: "m4.large"
        awsUser: "AWSPROGRAMMATICUSER"
        awsPassword: ${git.awspassword}
```
Cluster should be deployed by adding `.cluster.dev/minikube-one.yaml` into repository

In background there should be: 
 - Terraform module which creates Minikube instance with AWS Account credentials
 - Produced kubeconfig should be generated and passed to value into target git repo credentials
