# Adding a New Provider or Provisioner <!-- omit in toc -->

* [1. Create cluster manifest](#1-create-cluster-manifest)
* [2. Create Github Action and Gitlab pipeline](#2-create-github-action-and-gitlab-pipeline)
* [3. Define required treatment in main function](#3-define-required-treatment-in-main-function)
* [4. Define cloud related functions](#4-define-cloud-related-functions)
* [5. Create cleanup function](#5-create-cleanup-function)
* [6. Set cloud provisioner type](#6-set-cloud-provisioner-type)
* [7. Deploy mandatory addons](#7-deploy-mandatory-addons)
* [8. Test build](#8-test-build)

## 1. Create cluster manifest

Create a sample yaml cluster manifest with declaration for possible options in yaml and its naming.  
Naming for the options should be aligned with correspondent names in terraform provider or module.  
Set the list of options in [OPTIONS.md](OPTIONS.md).

## 2. Create Github Action and Gitlab pipeline

Defining required cloud authentication credentials, like Username/Password, Personal Access Tokens, or Access files.

## 3. Define required treatment in main function

Example:

```bash
# entrypoint.sh
    digitalocean)
        DEBUG "Cloud Provider: DigitalOcean"
        ;;
```

## 4. Define cloud related functions

Define cloud related functions in a dedicated file, for ex: `/bin/digitalocean_common.sh`.

Required functions:

| Function name       | What the function should do                                                |
| ------------------- | ---------------------------------------------------------------------- |
| `init_state_bucket` | check if the bucket exists and create a storage for terraform state                |
| `init_vpc`          | check and create required segmentation (this could be VPC, or Project) |
| `init_dns_zone`     | create a dns sub-zone which will be used for cluster services exposing |

## 5. Create cleanup function

Create a function that deletes:

* kubernetes cluster
* vpc/project
* domains

with dependent resources on yaml option `installed:false`

## 6. Set cloud provisioner type

The options on provisioners could differ even inside the same cloud provider, ex: Minikube, EKS, K3s.

The required set of functions is as follows:

| Function name        | What the function should do                                                           |
| -------------------- | --------------------------------------------------------------------------------- |
| `deploy_cluster`     | Deploy kubernetes cluster itself - save result in a separate tf-state               |
| `pull_kubeconfig`    | Obtain kubeconfig from the created cluster to use in the next steps with tf/helm          |
| `init_addons`        | Install mandatory kubernetes addons (see the next step) - write the result as a separate tf-state |
| `deploy_apps`        | Deploy/reconcile other applications from user's repository, from the defined folder    |
| `output_access_keys` | Add output with credentials URL's and other access parameters                     |

## 7. Deploy mandatory addons

Deploy mandatory addons for the cluster (`deploy_apps`):

* Storage Class for creating a PVC/PV
* Ingress Controller to serve traffic inside the cluster
* CertManager to create and manage certificates
* ExternalDNS to create DNS records
* ArgoCD to deploy an application

## 8. Test build

To test whether the build is successful, deploy a sample application from `/kubernetes/apps/samples/`.
