# Options List for Cluster Manifests


## Manifest Example
```yaml
# .cluster.dev/staging.yaml
cluster:
  installed: true
  name: staging
  cloud:
    provider: aws
    region: eu-central-1
    vpc: default
    domain: cluster.dev
    provisioner:
      type: minikube
      instanceType: m5.large
  addons:
    nginx-ingress: true
    cert-manager: true
  apps:
    - /kubernetes/apps/samples
```
More examples could be found in [/.cluster.dev](../.cluster.dev/) directory.

## Global Options

|  Key |  Required | Type  | Values  | Default  | Description |
|------|-----------|--------|---------|----------|----------------------------------------------|
|[name](#name)  | + | string | any     | | Cluster name to be used across all resources |
|[installed](#installed) | - | bool| `false`, `true`| `true`| Defines if cluster should be deployed or deleted, `false` would delete existing cluster |
|[cloud.provider](#cloud.provider)| + | string | `aws`, `digitalocean` | | Define cloud provider |
|[cloud.region](#cloud.region)| + | string|  ex: `us-east-1`, `do-fra-1` | | Define cloud provider region to create cluster |
|[cloud.domain](#cloud.domain)| - | string| FQDN ex: `cluster.dev`, `example.org` | `cluster.dev` | To expose cluster resources the DNS zone is required. If set to `cluster.dev` the installer would create a zone `cluster-name-organization.cluster.dev` and point it to your cloud service NS'es. Alternate you can set your own zone which already exist in target cloud.|
|[cloud.vpc](#cloud.vpc)| - |string|`default`,`create`,`vpc_id`| `default`| Virtual Private Cloud. `default` - use default one, `create` - installer would create a new VPC, `vpc_id` - define already existent.|


## Cluster Provisioners
### Amazon AWS Provisioners

Required Environment variables should be passed to container:  
`AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`. [AWS Docs](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_users_create.html)

|  Key |  Required | Type  | Values  | Default  | Description |
|------|-----------|--------|---------|----------|----------------------------------------------|
| [provisioner.type](#provisioner.type) | + | string | `minikube` | | Provisioner to deploy cluster with.|

#### AWS `minikube`
|  Key |  Required | Type  | Values  | Default  | Description |
|------|-----------|--------|---------|----------|----------------------------------------------|
| [provisioner.instanceType](#provisioner.instanceType) | + | string | ex:`m5.xlarge` | `m5.large` | Single node Kubernetes cluster AWS EC2 instance type. |


### DigitalOcean Provisioners

The DigitalOcean (DO) provider is used to interact with the resources supported by DigitalOcean.
Next environment variables should be set:  
`DIGITALOCEAN_TOKEN` - This is the DO API token. [DO Docs](https://www.digitalocean.com/docs/apis-clis/api/create-personal-access-token/)  
`SPACES_ACCESS_KEY_ID` - The access key ID used for Spaces API operations. [DO Docs](https://www.digitalocean.com/community/tutorials/how-to-create-a-digitalocean-space-and-api-key)  
`SPACES_SECRET_ACCESS_KEY` - The secret access key used for Spaces API operations.


|  Key |  Required | Type  | Values  | Default  | Description |
|------|-----------|--------|---------|----------|----------------------------------------------|
| [provisioner.type](#provisioner.type) | + | string | `managed-kubernetes` | | Provisioner to deploy cluster with.|
| [cloud.project](#cloud.project) | - | string | ex: `staging` | `default` | DigitalOcean Project name.|


#### DigitalOcean `managed-kubernetes`
|  Key |  Required | Type  | Values  | Default  | Description |
|------|-----------|--------|---------|----------|----------------------------------------------|
| [version](#version) | - | string | ex: `1.16` | | DigitalOcean managed Kubernetes [version](https://www.digitalocean.com/docs/kubernetes/changelog/). |
| [nodeCount](#nodeCount) | + | integer | `1-512`  | `1` | Number of Droplets instances in cluster. |
| [nodeSize](#nodeSize) | + | string | ex: `s-4vcpu-8gb`  | `s-1vcpu-2gb` | The slug identifier for the type of Droplet used as workers in the node pool. |
| [autoScale](#autoScale) | - | boolean | `true`, `false`  | `false` | A boolean indicating whether auto-scaling is enabled.|
| [minNodes](#minNodes) | - | boolean |`1-512` | `1` | If `autoScale` enabled defines a minimum number of Droplets instances in cluster. |
| [maxNodes](#maxNodes) | - | boolean |`1-512` | `1` | If `autoScale` enabled defines a maximum number of Droplets instances in cluster. |

## Cluster Addons
|  Key |  Required | Type  | Values  | Default  | Description |
|------|-----------|--------|---------|----------|----------------------------------------------|
| [nginx-ingress](#nginx-ingress) | - | boolean | `true`,`false` | `true` | Deploy [nginx-ingress](https://github.com/kubernetes/ingress-nginx). |
| [cert-manager](#cert-manager) | - | boolean | `true`,`false` | `true` | Deploy [cert-manager](https://cert-manager.io/). |
| [external-dns](#external-dns) | - | boolean | `true`,`false` | `true` | Deploy [external-dns](https://github.com/kubernetes-sigs/external-dns/). |
| [argo-cd](#argo-cd) | - | boolean | `true`,`false` | `true` | Deploy [argo-cd](https://argoproj.github.io/argo-cd/). |


# GIT Provider Support

## GitHub Actions Workflow Configuration

```yaml
# sample .github/workflows/aws.yaml
on:
  push:
# This how you can define after what changes it should be triggered
    paths:
      - '.cluster.dev/aws-minikube.yaml' 
    branches:
      - master
jobs:
  deploy_cluster_job:
    runs-on: ubuntu-latest
    name: Cluster.dev
    steps:
    - name: Checkout Repo
      uses: actions/checkout@v2
    - name: Reconcile Clusters
      id: reconcile
# Here you can define what release version of action to use,
# example: shalb/cluster.dev@master, shalb/cluster.dev@v0.1.7,  shalb/cluster.dev@test-branch
      uses: shalb/cluster.dev@v0.1.7
# Here the required environment variables should be set depending on Cloud Provider
      env:
        AWS_ACCESS_KEY_ID: "${{ secrets.AWS_ACCESS_KEY_ID }}"
        AWS_SECRET_ACCESS_KEY: "${{ secrets.AWS_SECRET_ACCESS_KEY }}"
        CLUSTER_CONFIG_PATH: "./.cluster.dev/"
# Here the debug level for the ACTION could be set (default: INFO)
        VERBOSE_LVL: DEBUG
    - name: Get the Cluster Credentials
      run: echo -e "\n\033[1;32m${{ steps.reconcile.outputs.ssh }}\n\033[1;32m${{ steps.reconcile.outputs.kubeconfig }}\n\033[1;32m${{ steps.reconcile.outputs.argocd }}"
```
More examples could be found in [/.github/workflows](../.github/workflows) directory.

## GitLab CI/CD Pipeline Configuration

```yaml
# Example for .gitlab-ci.yml pipeline with cluster.dev job
image: docker:19.03.0

variables:
  DOCKER_DRIVER: overlay2 # Docker Settings
  DOCKER_TLS_CERTDIR: "/certs"
  CLUSTER_DEV_BRANCH: "master" # Define branch or release version
  CLUSTER_CONFIG_PATH: "./.cluster.dev/" # Path to manifests
  DIGITALOCEAN_TOKEN: "${DIGITALOCEAN_TOKEN}"  # Environment variables depending on Cloud Provider
  SPACES_ACCESS_KEY_ID: "${SPACES_ACCESS_KEY_ID}"
  SPACES_SECRET_ACCESS_KEY: "${SPACES_SECRET_ACCESS_KEY}"

services:
  - docker:19.03.0-dind

before_script:
  - apk update && apk upgrade && apk add --no-cache bash git

stages:
  - cluster-dev

cluster-dev:
  only:
    refs:
      - master
    changes:
      - '.gitlab-ci.yml'
      - '.cluster.dev/**' # Path to cluster declaration manifests
      - '/kubernetes/apps/**' # ArgoCD application directories
  script:
    - git clone -b "$CLUSTER_DEV_BRANCH" https://github.com/shalb/cluster.dev.git
    - cd cluster.dev && docker build --no-cache -t "cluster.dev" .
    - docker run --name cluster.dev --workdir /gitlab/workspace --rm -e CI_PROJECT_PATH -e CI_PROJECT_DIR -e VERBOSE_LVL=DEBUG -e DIGITALOCEAN_TOKEN -e SPACES_ACCESS_KEY_ID -e SPACES_SECRET_ACCESS_KEY -v "${CI_PROJECT_DIR}:/gitlab/workspace" cluster.dev
  stage: cluster-dev
```
Full example could be found in [/install/.gitlab-ci-sample.yml](../install/.gitlab-ci-sample.yml)

