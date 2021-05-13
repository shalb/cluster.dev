# Download Sample Project

cdev uses [project templates](https://cluster.dev/template-development/) to generate users' projects in a desired cloud. Given below is the description of three available templates that are ready to use.

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

### Quick start

1. [Configure access to AWS](https://cluster.dev/aws-cloud-provider/) and export required variables.

2. Clone example project:

    ```bash
      git clone https://github.com/shalb/cdev-aws-k3s.git
      cd cdev-aws-k3s/examples/
    ```

3. Edit variables in the example's files, if necessary:

    * project.yaml - main project config. Sets common global variables for current project such as organization, region, state bucket name etc. See [project configuration docs](https://cluster.dev/project-configuration/#project).

    * backend(s).yaml - configures backend for cdev states (including Terraform states). Uses variables from project.yaml. See [backend docs](https://cluster.dev/project-configuration/#backends).

    * infrastructure.yaml - describes infrastructure configuration. See [infrastructure docs](https://cluster.dev/project-configuration/#infrastructure).

4. Run `cdev plan`

5. Run `cdev apply` . We highly recommend to run `cdev apply` in a debug mode so that you could see cdev logging in the output: `cdev apply -l debug`

6. After `cdev apply` is successfully executed, in the output you will see the ArgoCD URL of your cluster. Sign in to the console to check whether ArgoCD is up and running and the template has been deployed correctly. To sign in, use the "admin" login and the bcrypted password that you have generated for the infrastructure.yaml.

7. Displayed in the output will be also a command on how to get kubeconfig and connect to your Kubernetes cluster.

8. Shut down the cluster with the command `cdev destroy`

9. Alternatively, you can also use [code generator](https://cluster.dev/quick-start/) to create the same example.

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

### Quick start

1. [Configure access to AWS](https://cluster.dev/aws-cloud-provider/) and export required variables.

2. Clone example project:

    ```bash
      git clone https://github.com/shalb/cdev-aws-eks.git
      cd cdev-aws-eks/examples/
    ```

3. Edit variables in the example's files, if necessary:

    * project.yaml - main project config. Sets common global variables for current project such as organization, region, state bucket name etc. See [project configuration docs](https://cluster.dev/project-configuration/#project).

    * backend(s).yaml - configures backend for cdev states (including Terraform states). Uses variables from project.yaml. See [backend docs](https://cluster.dev/project-configuration/#backends).

    * infrastructure.yaml - describes infrastructure configuration. See [infrastructure docs](https://cluster.dev/project-configuration/#infrastructure).

4. Run `cdev plan`

5. Run `cdev apply` . We highly recommend to run `cdev apply` in a debug mode so that you could see cdev logging in the output: `cdev apply -l debug`

6. After `cdev apply` is successfully executed, in the output you will see the ArgoCD URL of your cluster. Sign in to the console to check whether ArgoCD is up and running and the template has been deployed correctly. To sign in, use the "admin" login and the bcrypted password that you have generated for the infrastructure.yaml.

7. Displayed in the output will be also a command on how to get kubeconfig and connect to your Kubernetes cluster.

8. Shut down the cluster with the command `cdev destroy`

9. Alternatively, you can also use [code generator](https://cluster.dev/quick-start/) to create the same example.

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

### Quick Start

1. [Configure access to DO](https://cluster.dev/digital-ocean-cloud-provider/) and export required variables.

2. Clone example project:

    ```bash
    git clone https://github.com/shalb/cdev-do-k8s.git
    cd cdev-do-k8s/examples/
    ```

3. Edit variables in the example's files, if necessary:

    * project.yaml - main project config. Sets common global variables for current project such as organization, region, state bucket name etc. See [project configuration docs](https://cluster.dev/project-configuration/#project).

    * backend(s).yaml - configures backend for cdev states (including Terraform states). Uses variables from project.yaml. See [backend docs](https://cluster.dev/project-configuration/#backends).

    * infrastructure.yaml - describes infrastructure configuration. See [infrastructure docs](https://cluster.dev/project-configuration/#infrastructure).

4. Run `cdev plan`

5. Run `cdev apply` . We highly recommend to run `cdev apply` in a debug mode so that you could see cdev logging in the output: `cdev apply -l debug`

6. After `cdev apply` is successfully executed, in the output you will see the ArgoCD URL of your cluster. Sign in to the console to check whether ArgoCD is up and running and the template has been deployed correctly. To sign in, use the "admin" login and the bcrypted password that you have generated for the infrastructure.yaml.

7. Displayed in the output will be also a command on how to get kubeconfig and connect to your Kubernetes cluster.

8. Shut down the cluster with the command `cdev destroy`

9. Alternatively, you can also use [code generator](https://cluster.dev/quick-start/) to create the same example.
