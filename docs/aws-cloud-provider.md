# Deploying to AWS

cdev uses stack templates to generate projects in a desired cloud. This section describes steps necessary to start working with cdev in AWS cloud using [AWS-EKS](https://github.com/shalb/cdev-aws-eks) stack template.

## Prerequisites to use AWS-EKS stack template

1. Terraform version 13+.

2. AWS account.

3. [AWS CLI](#install-aws-client-and-check-access) installed.

4. kubectl installed.

5. [cdev installed](https://cluster.dev/installation/).

### Authentication

cdev requires cloud credentials to manage and provision resources. You can configure access to AWS in two ways:

!!! Info
    Please note that you have to use IAM user with granted administrative permissions.

* **Environment variables**: provide your credentials via the `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY`, the environment variables that represent your AWS Access Key and AWS Secret Key. You can also use the `AWS_DEFAULT_REGION` or `AWS_REGION` environment variable to set region, if needed. Example usage:

    ```bash
    export AWS_ACCESS_KEY_ID="MYACCESSKEY"
    export AWS_SECRET_ACCESS_KEY="MYSECRETKEY"
    export AWS_DEFAULT_REGION="eu-central-1"
    ```

* **Shared Credentials File (recommended)**: set up an [AWS configuration file](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-files.html) to specify your credentials.

    Credentials file `~/.aws/credentials` example:

    ```bash
    [cluster-dev]
    aws_access_key_id = MYACCESSKEY
    aws_secret_access_key = MYSECRETKEY
    ```

    Config: `~/.aws/config` example:

    ```bash
    [profile cluster-dev]
    region = eu-central-1
    ```

    Then export `AWS_PROFILE` environment variable.

    ```bash
    export AWS_PROFILE=cluster-dev
    ```

### Install AWS client and check access

If you do not have the AWS CLI installed, refer to AWS CLI [official installation guide](https://docs.aws.amazon.com/cli/latest/userguide/install-cliv2-linux.html), or use commands from the example:

```bash
curl "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv2.zip"
unzip awscliv2.zip
sudo ./aws/install
aws s3 ls
```

### Create S3 bucket for states

cdev uses S3 bucket for storing states. Create the bucket with the command:

```bash
aws s3 mb s3://cdev-states
```

### DNS Zone

In AWS-EKS stack template example you need to define a Route 53 hosted zone. Options:

1. You already have a Route 53 hosted zone.

2. Create a new hosted zone using a [Route 53 documentation example](https://docs.aws.amazon.com/cli/latest/reference/route53/create-hosted-zone.html#examples).

3. Use "cluster.dev" domain for zone delegation.

## AWS-EKS

[AWS-EKS](https://github.com/shalb/cdev-aws-eks) is a cdev stack template that creates and provisions Kubernetes clusters in AWS cloud by means of Amazon Elastic Kubernetes Service (EKS).

### AWS-EKS starting guide

1. Configure [access to AWS](#authentication) and export required variables.

2. Create locally a project directory, cd into it and execute the command:

    ```bash
      cdev project create https://github.com/shalb/cdev-aws-eks
    ```

    !!! tip
        The stack template's repo could contain several options for project generation. To list available generators, use ```--list-templates``` option:

          ```bash
          cdev project create https://github.com/shalb/cdev-aws-eks --list-templates
          ```

        Then you can specify which generator to use, for example:

          ```bash
          cdev project create https://github.com/shalb/cdev-aws-eks minimal
          ```
    !!! tip
        If you leave it unspecified, cdev will generate a default project for you. You can also opt for an  interactive mode with the extended menu:

          ```bash
          cdev project create https://github.com/shalb/cdev-aws-eks --interactive
          ```

3. Edit variables in the example's files, if necessary:

    * project.yaml - main project config. Sets common global variables for current project such as organization, region, state bucket name etc. See [project configuration docs](https://cluster.dev/project-configuration/#project).

    * backend.yaml - configures backend for cdev states (including Terraform states). Uses variables from project.yaml. See [backend docs](https://cluster.dev/project-configuration/#backends).

    * infra.yaml - describes stack configuration. See [infrastructure docs](https://cluster.dev/project-configuration/#infrastructure).

4. Run `cdev plan` to build the project. In the output you will see an infrastructure that is going to be created after running `cdev apply`.

    !!! note
        Prior to running `cdev apply` make sure to look through the infra.yaml file and replace the commented fields with real values. In case you would like to use existing VPC and subnets, uncomment preset options and set correct VPC ID and subnets' IDs. If you leave them as is, cdev will have VPC and subnets created for you.

5. Run `cdev apply`

    !!! tip
        We highly recommend to run `cdev apply` in a debug mode so that you could see cdev logging in the output: `cdev apply -l debug`

6. After `cdev apply` is successfully executed, in the output you will see the ArgoCD URL of your cluster. Sign in to the console to check whether ArgoCD is up and running and the stack template has been deployed correctly. To sign in, use the "admin" login and the bcrypted password that you have generated for the infra.yaml.

7. Displayed in the output will be also a command on how to get kubeconfig and connect to your Kubernetes cluster.

8. Destroy the cluster and all created resources with the command `cdev destroy`

### Resources to be created

* *(optional, if you use cluster.dev domain)* Route53 zone **<cluster-name>.cluster.dev**

* *(optional, if vpc_id is not set)* VPC for EKS cluster

* EKS Kubernetes cluster with addons:

    * cert-manager

    * ingress-nginx

    * external-dns

    * argocd

* AWS IAM roles for EKS IRSA cert-manager and external-dns
