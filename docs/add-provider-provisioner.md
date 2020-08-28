---
layout: default
title: "Adding New Provider or Provisioner to Cluster.dev"
permalink: /add-provider-provisioner/
---
1. Create a sample yaml cluster manifest with declaration for possible options in yaml and its naming. Naming for the options should be aligned with correspondent names in terraform provider or module. Set the list of options in [OPTIONS.md](OPTIONS.md).

2. Create a sample GitHub Action and GitLab pipeline defining required cloud authentication credentials, like Username/Password, Personal Access Tokens, or Access files.

3. Define required treatment in main function, ex:
```yaml
# entrypoint.sh
    digitalocean)
        DEBUG "Cloud Provider: DigitalOcean"
        ;;
```
4. Define cloud related functions in a dedicated file, ex:`/bin/digitalocean_common.sh`, required functions:
    - `init_state_bucket`  # Check if exists and create a storage for terraform state
    - `init_vpc`           # Check and create required segmentation (this could be VPC, or Project)
    - `init_dns_zone`      # Create a dns sub-zone that will be used for cluster services exposing.

5. Create a function that deletes: Kubernetes cluster, VPC/project, domains, with dependent resources on yaml option `installed:false`.

6. Set the provisioner type inside the cloud. The options on provisioners could differ even inside the same cloud provider, ex: Minikube, EKS, K3s. The required set of functions is as follows:
    - `deploy_cluster`      # Deploys Kubernetes cluster itself - saves the result in a separate tf-state
    - `pull_kubeconfig`     # Obtains kubeconfig from the created cluster to use in the next steps with tf/helm
    - `init_addons`         # Installs mandatory Kubernetes addons (see the next step) - writes the result as a separate tf-state
    - `deploy_apps`         # Deploys/reconciles other applications from user's repository, from the defined folder
    - `output_access_keys`  # Adds output with credentials URL's and other access parameters.

7. Deploy a mandatory addons for cluster (`deploy_apps`):
    - Storage Class for creating a PVC/PV
    - Ingress Controller to serve traffic inside the cluster
    - CertManager to create and manage certificates
    - ExternalDNS to create DNS records
    - ArgoCD to deploy an application

8. To test whether the build is successful, deploy a sample application from `/kubernetes/apps/samples/`
