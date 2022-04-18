* [Info](#info)
* [Requirements](#requirements)
   * [Set variables regarding to your needs](#set-variables-regarding-to-your-needs)
      * [Create env file](#create-env-file)
      * [Customize the enviroment file](#customize-the-enviroment-file)
      * [Export all variables](#export-all-variables)
   * [Install terraform](#install-terraform)
   * [Install cluster.dev](#install-clusterdev)
   * [Install doctl](#install-doctl)
   * [Install kubectl](#install-kubectl)
   * [Create DO's space to store your project's state](#create-dos-space-to-store-your-projects-state)
      * [Configure access to the space](#configure-access-to-the-space)
   * [Configure DO's API access via token](#configure-dos-api-access-via-token)
* [Create basic project](#create-basic-project)
   * [Customize project](#customize-project)
* [Create new cluster by the cdev](#create-new-cluster-by-the-cdev)
* [Destroy the cluster](#destroy-the-cluster)

# Info

We will learn how to run test Kubernetes cluster in [DigitalOcean](https://www.digitalocean.com/) by [Clusterdev](https://cluster.dev/)

# Requirements

You should have any [Ubuntu 20.04](https://releases.ubuntu.com/20.04/) host to use this manual without any customization

## Set variables regarding to your needs
### Create env file
Create default enviroment file:
```bash
echo \
'export PROJECT="example"
# Space access token:
export SPACES_ACCESS_KEY_ID=
export SPACES_SECRET_ACCESS_KEY=
# API access token:
export DIGITALOCEAN_TOKEN=
# Generic default settings:
export REGION="fra1"
export BUCKED_NAME="${PROJECT}-cdev-state"
export ORGANIZATION="${PROJECT}-organization"
export CLUSTER_NAME="${PROJECT}-cdev"
export NODE_TYPE="s-4vcpu-8gb-intel"
export ARCH="amd64"
export TERRAFORM_VERSION="1.1.8"
export CDEV_VERSION="v0.6.3"
export DOCTL_VERSION="1.72.0"
export KUBECTL_VERSION="v1.22.8"
export KUBECONFIG="/tmp/kubeconfig_${CLUSTER_NAME}"
' > project_env
```
### Customize the enviroment file
You should change at least `PROJECT` environment variable to make it unique:
```bash
editor project_env
```

### Export all variables
```bash
source project_env
```

## Install terraform
[Terraform](https://www.terraform.io/) needed to deploy our infrastracture
```bash
apt install -y unzip
curl -O https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_linux_${ARCH}.zip
unzip terraform_${TERRAFORM_VERSION}_linux_${ARCH}.zip
sudo mv terraform /usr/local/bin/
sudo chown root:root /usr/local/bin/terraform
sudo chmod 755 /usr/local/bin/terraform
rm terraform_${TERRAFORM_VERSION}_linux_${ARCH}.zip
terraform --version
```

## Install cluster.dev
[Clusterdev](https://cluster.dev/) needed to simplify whole process of infrastracture deployment
```bash
wget https://github.com/shalb/cluster.dev/releases/download/${CDEV_VERSION}/cdev-${CDEV_VERSION}-linux-${ARCH}.tar.gz
tar -xzf cdev-${CDEV_VERSION}-linux-${ARCH}.tar.gz
sudo mv cdev /usr/local/bin/
sudo chown root:root /usr/local/bin/cdev
sudo chmod 755 /usr/local/bin/cdev
rm cdev-${CDEV_VERSION}-linux-${ARCH}.tar.gz
cdev --version
```

## Install doctl
[Doctl](https://github.com/digitalocean/doctl) used by the Clusterdev and will help to interact with the DO from command line
```bash
wget https://github.com/digitalocean/doctl/releases/download/v${DOCTL_VERSION}/doctl-${DOCTL_VERSION}-linux-${ARCH}.tar.gz
tar -xzf doctl-${DOCTL_VERSION}-linux-${ARCH}.tar.gz
sudo mv doctl /usr/local/bin/
sudo chown root:root /usr/local/bin/doctl
sudo chmod 755 /usr/local/bin/doctl
rm doctl-${DOCTL_VERSION}-linux-${ARCH}.tar.gz
doctl version
```

## Install kubectl
[Kubectl](https://kubernetes.io/ru/docs/tasks/tools/install-kubectl/) needed to interact with kubernetes cluster
```bash
curl -LO https://storage.googleapis.com/kubernetes-release/release/v1.22.8/bin/linux/amd64/kubectl
chmod 755 kubectl
sudo mv kubectl /usr/local/bin/kubectl
kubectl version
```

## Create DO's space to store your project's state
Go to [spaces](https://cloud.digitalocean.com/spaces/) and create new space.  
Use name produced by this echo command:
```bash
echo ${BUCKED_NAME}
```

### Configure access to the space 
Go to [API tokens](https://cloud.digitalocean.com/account/api/tokens)(**Spaces access keys** section) and create new token to allow the Clusterdev to use the storage as Terraform's state location  
Add storage token to the enviroment file:
```bash
editor project_env
```

## Configure DO's API access via token
Go to [API tokens](https://cloud.digitalocean.com/account/api/tokens)(**Personal access tokens** section) and create new token to allow the Clusterdev to manage infrastructure  
Add API token to the enviroment file:
```bash
editor project_env
```

# Create basic project
```bash
mkdir cdev && mv project_env cdev/ && cd cdev
cdev project create https://github.com/shalb/cdev-do-k8s?ref=v0.0.2
```

## Customize project
```bash
source project_env

echo \
"name: demo-project
kind: Project
variables:
  organization: ${ORGANIZATION}
  region: ${REGION}
  bucket_name: ${BUCKED_NAME}
" > project.yaml

sed -i 's/"s-1vcpu-2gb"/{{ reqEnv "NODE_TYPE" }}/g' stack.yaml
sed -i 's/arti-k8s/{{ reqEnv "CLUSTER_NAME" }}/g' stack.yaml
```

# Create new cluster by the cdev
```bash
tmux
cdev apply -l debug | tee log
```

# Destroy the cluster
```bash
cdev destroy -l debug | tee log
```
