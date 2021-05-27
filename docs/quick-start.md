# Quick Start

cdev uses [project templates](https://cluster.dev/template-development/) to generate users' projects in a desired cloud. In this section you will find a quick guidance on how to create your first project, and the description of three available templates that are ready to use.

## Step-by-step guide

This guide explains how to create and deploy your first project using cdev.

Before you begin make sure that you have [cdev installed](https://cluster.dev/installation/) and comply with all [preconditions](https://cluster.dev/prerequisites/) necessary to start using it.

1. Configure access to your desired cloud ([AWS](https://cluster.dev/aws-cloud-provider/),[DigitalOcean](https://cluster.dev/digital-ocean-cloud-provider/), [GCE](https://cluster.dev/google-cloud-provider/), [Azure](https://cluster.dev/azure-cloud-provider/) and export required variables.

2. Create locally a project directory, cd into it and execute the command depending on a chosen template. For AWS-EKS:

    ```bash
      cdev project create https://github.com/shalb/cdev-aws-eks
    ```

    For AWS-k3s:

    ```bash
      cdev project create https://github.com/shalb/cdev-aws-k3s
    ```

    For DO-k8s:

    ```bash
      cdev project create https://github.com/shalb/cdev-do-k8s
    ```

    As the template's repo could contain several options for project generation, you can specify which generator to use, for example:

    ```bash
      cdev project create https://github.com/shalb/cdev-aws-eks minimal
    ```

    If you leave it unspecified, cdev will generate a default project for you. You can also opt for an interactive mode with the extended menu:

    ```bash
      cdev project create https://github.com/shalb/cdev-aws-eks --interactive
    ```

3. Edit variables in the example's files, if necessary:

    * project.yaml - main project config. Sets common global variables for current project such as organization, region, state bucket name etc. See [project configuration docs](https://cluster.dev/project-configuration/#project).

    * backend.yaml - configures backend for cdev states (including Terraform states). Uses variables from project.yaml. See [backend docs](https://cluster.dev/project-configuration/#backends).

    * infrastructure.yaml - describes infrastructure configuration. See [infrastructure docs](https://cluster.dev/project-configuration/#infrastructure).

4. Run `cdev plan` to build the project. In the output you will see an infrastructure that is going to be created after running `cdev apply`.

    Prior to running `cdev apply` make sure to look through the infrastructure.yaml file and replace the commented fields with real values. In case you would like to use existing VPC and subnets, uncomment preset options and set correct VPC ID and subnets' IDs. If you leave them as is, cdev will have VPC and subnets created for you.

5. Run `cdev apply` . We highly recommend to run `cdev apply` in a debug mode so that you could see cdev logging in the output: `cdev apply -l debug`

6. After `cdev apply` is successfully executed, in the output you will see the ArgoCD URL of your cluster. Sign in to the console to check whether ArgoCD is up and running and the template has been deployed correctly. To sign in, use the "admin" login and the bcrypted password that you have generated for the infrastructure.yaml.

7. Displayed in the output will be also a command on how to get kubeconfig and connect to your Kubernetes cluster.

8. Destroy the cluster and all created resources with the command `cdev destroy`

## AWS-EKS

[AWS-EKS is a cdev template](https://github.com/shalb/cdev-aws-eks) that creates and provisions Kubernetes clusters in [AWS cloud](https://cluster.dev/aws-cloud-provider/) by means of Amazon Elastic Kubernetes Service (EKS). The resources to be created:

* *(optional, if your use cluster.dev domain)* Route53 zone **<cluster-name>.cluster.dev**

* *(optional, if vpc_id is not set)* VPC for EKS cluster

* EKS Kubernetes cluster with addons:

    * cert-manager

    * ingress-nginx

    * external-dns

    * argocd

* AWS IAM roles for EKS IRSA cert-manager and external-dns.

### Prerequisites

1. Terraform version 13+.

2. AWS account.

3. AWS CLI installed.

4. kubectl installed.

5. [cdev installed](https://cluster.dev/installation/).

## AWS-k3s

[AWS-k3s is a cdev template](https://github.com/shalb/cdev-aws-k3s) that creates and provisions Kubernetes clusters in [AWS cloud](https://cluster.dev/aws-cloud-provider/) by means of k3s utility. The template deploys a k3s cluster using existing or created VPC and domain name. The resources to be created:

* AWS Key Pair to access running instances of the cluster.

* *(optional, if your use cluster.dev domain)* Route53 zone **<cluster-name>.cluster.dev**

* *(optional, if vpc_id is not set)* VPC for EKS cluster

* AWS IAM Policy for managing your DNS zone by external-dns

* k3s Kubernetes cluster with addons:

    * cert-manager

    * ingress-nginx

    * external-dns

    * argocd

### Prerequisites

1. Terraform version 13+.

2. AWS account.

3. AWS CLI installed.

4. kubectl installed.

5. [cdev installed](https://cluster.dev/installation/).

## DO-k8s

[DO-k8s is a cdev template](https://github.com/shalb/cdev-do-k8s) that creates and provisions Kubernetes clusters in the DigitalOcean cloud. The resources to be created:

* *(optional, if vpc_id is not set)* VPC for Kubernetes cluster
* DO Kubernetes cluster with addons:
    * cert-manager
    * argocd

### Prerequisites

1. Terraform version 13+.

2. DigitalOcean account.

3. [doctl installed](https://docs.digitalocean.com/reference/doctl/how-to/install/).

4. [cdev installed](https://cluster.dev/installation/).
