---
layout: default
title: "Adding New Provider or Provisioner to Cluster.dev"
permalink: /add-provider-provisioner/
---

## Adding a New Provider or Provisioner
1. Create a sample yaml cluster manifest with declaration for possible options in yaml and its naming.  
   Naming for the options should be aligned with correspondent names in terraform provider or module.

2. Create a sample Github Action and Gitlab pipeline defining required cloud authentication credentials, like Username/Password, Personal Access Tokens, or Access files, ex:
```yaml
    # For user and password please use your token name and token hash
    # https://www.digitalocean.com/docs/apis-clis/api/create-personal-access-token/
        cloud-user: ${{ secrets.DO_TOKEN_NAME }}
        cloud-pass: ${{ secrets.DIGITALOCEAN_TOKEN }}
```
3. Define required treatment in main function, ex:
```yaml
# entrypoint.sh
    digitalocean)
        DEBUG "Cloud Provider: DigitalOcean"
        ;;
```
4. Define cloud related functions in a dedicated file, ex:`/bin/digitalocean_common.sh`, required functions:
    - `init_state_bucket`  # check if exist and create a storage for terraform state 
    - `init_vpc`           # check and create required segmentation (this could be VPC, or Project)
    - `init_dns_zone`      # create a dns sub-zone which will be used for cluster services exposing

5. Create a functions which deletes: kubernetes cluster, vpc/project, domains, with dependent resources on yaml option `installed:false`

6. Set provisioner type inside the cloud. The options on provisioners could different even inside same cloud provider ex: Minikube, EKS, K3s. The required set of functions are next:
    - `deploy_cluster`      # Deploy kubernetes cluster itself - save result in separate tf-state
    - `pull_kubeconfig`     # Obtain kubeconfig from created cluster to use in next steps with tf/helm
    - `init_addons`         # Install mandatory kubernetes addons (see next step) - result as separate tf-state
    - `deploy_apps`         # Deploy/reconcile other applications from user's repository from defined folder
    - `output_access_keys`  # Add output with credentials URL's and other access parameters

7. Deploy a mandatory addons for cluster (`deploy_apps`):
    - Storage Class for creating a PVC/PV
    - Ingress Controller to serve traffic inside cluster
    - CertManager to create and manage certificates
    - ExternalDNS to create DNS records
    - ArgoCD to deploy an applications

8. To test successful build - deploy a sample applications from `/kubernetes/apps/samples/`
