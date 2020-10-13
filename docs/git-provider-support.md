
# Git Provider Support

## GitHub Actions Workflow Configuration

*More examples could be found in [/.github/workflows](https://github.com/shalb/cluster.dev/tree/master/.github/workflows) directory.*

```yaml
# sample .github/workflows/aws.yaml
on:
  push:
# This is how you can define after what changes it should be triggered
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
# example: shalb/cluster.dev@master, shalb/cluster.dev@v0.3.3,  shalb/cluster.dev@test-branch
      uses: shalb/cluster.dev@v0.3.3
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

## GitLab CI/CD Pipeline Configuration

*Full example could be found in [/install/.gitlab-ci-sample.yml](https://github.com/shalb/cluster.dev/blob/master/install/.gitlab-ci-sample.yml)*


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
