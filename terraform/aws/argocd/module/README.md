## Terraform module to deploy ArgoCD to Kubernetes cluster
For input is should receive Kubernetes kubeconfig,   
or hash from it: `md5sum ~/.kube/config | cut -c 1-8`  
Hash should be used for resource uniqness.
In case of cluster changed, resources should be re-deployed.