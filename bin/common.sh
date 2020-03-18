#!/usr/bin/env bash

source ./logging.sh

#######################################
# Create or use exiting S3 bucket for Terraform states
# Globals:
#   S3_BACKEND_BUCKET
# Arguments:
#   cluster_cloud_region
# Outputs:
#   Writes progress status
#######################################
function aws::init_s3_bucket {
    DEBUG "Create or use exiting S3 bucket for Terraform states"
    local cluster_cloud_region=$1

    cd terraform/aws/backend/ || ERROR "Path not found"

    # Create and init backend.
    run_cmd "terraform init"

    # Check if bucket already exist by trying to import it
    if (terraform import -var="region=$cluster_cloud_region" -var="s3_backend_bucket=$S3_BACKEND_BUCKET" aws_s3_bucket.terraform_state "$S3_BACKEND_BUCKET" >/dev/null 2>&1); then
        INFO "Terraform S3_BACKEND_BUCKET: $S3_BACKEND_BUCKET already exist"
    else
        NOTICE "Terraform S3_BACKEND_BUCKET: $S3_BACKEND_BUCKET not exist. It is going to be created"
        run_cmd "terraform apply -auto-approve \
                    -var='region=$cluster_cloud_region' \
                    -var='s3_backend_bucket=$S3_BACKEND_BUCKET'"
    fi

    cd - >/dev/null || ERROR "Path not found"
}

#######################################
# Create a DNS domains/records if required
# TODO: implement switch for domain. https://github.com/shalb/cluster.dev/issues/2
# Globals:
#   S3_BACKEND_BUCKET
#   CLUSTER_FULLNAME
# Arguments:
#   cluster_cloud_region
#   cluster_name
#   cluster_cloud_domain
# Outputs:
#   Writes progress status
#######################################
function aws::init_route53 {
    DEBUG "Create a DNS domains/records if required"
    local cluster_cloud_region=$1
    local cluster_name=$2
    local cluster_cloud_domain=$3

    cd terraform/aws/route53/ || ERROR "Path not found"

    if [ -z "$cluster_cloud_domain" ]; then
        INFO "The cluster domain is unset. It is going to be created default"
    else
        INFO "The cluster domain is defined. So applying Terraform configuration for it"
    fi

    # terraform init -backend-config="bucket=$S3_BACKEND_BUCKET" \
    #     -backend-config="key=$cluster_name/terraform.state" \
    #     -backend-config="region=$cluster_cloud_region"
    # terraform plan -compact-warnings \
    #     -var="region=$cluster_cloud_region" \
    #     -var="cluster_fullname=$CLUSTER_FULLNAME" \
    #     -var="cluster_domain=$cluster_cloud_domain"

    cd - >/dev/null || ERROR "Path not found"
}

#######################################
# Create a VPC or use existing defined
# Globals:
#   S3_BACKEND_BUCKET
#   CLUSTER_FULLNAME
# Arguments:
#   cluster_cloud_vpc
#   cluster_name
#   cluster_cloud_region
# Outputs:
#   Writes progress status
# KEY: vpc (cluster_cloud_vpc)
# Possible options:
#   default - use default vpc subnet
#   create - create new vpc by terraform
#   vpc-id - use client vpc, first subnet in a list
#######################################
function aws::init_vpc {
    DEBUG "Create a VPC or use existing defined"
    local cluster_cloud_vpc=$1
    local cluster_cloud_vpc_id=""

    cd terraform/aws/vpc/ || ERROR "Path not found"

    case ${cluster_cloud_vpc} in
        default|"")
            INFO "Use default VPC"
            ;;
        create)
            # Create new VPC and get ID.
            NOTICE "Creating new VPC"
            INFO "VPC: Initializing Terraform configuration"
            run_cmd "terraform init \
                        -backend-config='bucket=$S3_BACKEND_BUCKET' \
                        -backend-config='key=$cluster_name/terraform-vpc.state' \
                        -backend-config='region=$cluster_cloud_region'"

            run_cmd "terraform plan \
                        -var='region=$cluster_cloud_region' \
                        -var='cluster_name=$CLUSTER_FULLNAME' \
                        -input=false \
                        -out=tfplan"

            INFO "VPC: Creating infrastructure"
            run_cmd "terraform apply -auto-approve -compact-warnings -input=false tfplan"
            # Get VPC ID for later use.
            cluster_cloud_vpc_id=$(terraform output vpc_id)
            ;;
        *)
            # Use client VPC ID.
            INFO "VPC ID in use: ${cluster_cloud_vpc}"
            cluster_cloud_vpc_id=${cluster_cloud_vpc}
            ;;
    esac

    cd - >/dev/null || ERROR "Path not found"
}

#######################################
# Pull a kubeconfig to instance via kubectl
# Globals:
#   CLUSTER_FULLNAME
# Arguments:
#   None
# Outputs:
#   Writes progress status
#######################################
function aws::minikube::pull_kubeconfig {
    DEBUG "Pull a kubeconfig to instance via kubectl"
    local WAIT_TIMEOUT=5

    INFO "Copy kubeconfig to instance with Minikube"
    export KUBECONFIG=~/.kube/kubeconfig_${CLUSTER_FULLNAME}

    INFO "Waiting for the Kubernetes Cluster to get ready. It can take some time"
    until kubectl version --request-timeout=5s >/dev/null 2>&1; do
        DEBUG "Waiting ${WAIT_TIMEOUT}s"
        sleep $WAIT_TIMEOUT

        run_cmd "aws s3 cp 's3://${CLUSTER_FULLNAME}/kubeconfig_$CLUSTER_FULLNAME' '$HOME/.kube/kubeconfig_$CLUSTER_FULLNAME' 2>/dev/null"
        run_cmd "cp '$HOME/.kube/kubeconfig_$CLUSTER_FULLNAME' '$HOME/.kube/config' 2>/dev/null"
    done
}

#######################################
# Deploy Minikube cluster via Terraform
# Globals:
#   S3_BACKEND_BUCKET
#   CLUSTER_FULLNAME
# Arguments:
#   cluster_name
#   cluster_cloud_region
#   cluster_provisioner_instanceType
#   cluster_cloud_domain
# Outputs:
#   Writes progress status
#######################################
function aws::minikube::deploy_cluster {
    DEBUG "Deploy Minikube cluster via Terraform"
    local cluster_name=$1
    local cluster_cloud_region=$2
    local cluster_provisioner_instanceType=$3
    local cluster_cloud_domain=$4

    cd terraform/aws/minikube/ || ERROR "Path not found"

    # Deploy main Terraform code
    INFO "Minikube cluster: Initializing Terraform configuration"
    run_cmd "terraform init \
                -backend-config='bucket=$S3_BACKEND_BUCKET' \
                -backend-config='key=$cluster_name/terraform.state' \
                -backend-config='region=$cluster_cloud_region'"

    # TODO: Minikube module is using Centos7 image which requires to be accepted and subscribed in MarketPlace: https://github.com/shalb/cluster.dev/issues/9
    # To do so please visit https://aws.amazon.com/marketplace/pp?sku=aw0evgkw8e5c1q413zgy5pjce

    run_cmd "terraform plan \
                -var='region=$cluster_cloud_region' \
                -var='cluster_name=$CLUSTER_FULLNAME' \
                -var='aws_instance_type=$cluster_provisioner_instanceType' \
                -var='hosted_zone=$cluster_cloud_domain' \
                -input=false \
                -out=tfplan"

    INFO "Minikube cluster: Creating infrastructure"
    run_cmd "terraform apply -auto-approve -compact-warnings -input=false tfplan"

    cd - >/dev/null || ERROR "Path not found"
}

#######################################
# Deploy CertManager via kubectl
# Globals:
#   None
# Arguments:
#   None
# Outputs:
#   Writes progress status
#######################################
function deploy_cert_manager {
    DEBUG "Deploy CertManager via kubectl"

    INFO "Setup TLS certificates"
    run_cmd "kubectl apply -f 'https://github.com/jetstack/cert-manager/releases/download/v0.13.0/cert-manager-no-webhook.yaml'"
    run_cmd "kubectl apply -f 'https://raw.githubusercontent.com/shalb/terraform-aws-minikube/8a147f7c0044c318ec37990b50f0cabb205e9b44/addons/letsencrypt-prod.yaml'"
}

#######################################
# Deploy ArgoCD via Terraform
# Globals:
#   S3_BACKEND_BUCKET
#   CLUSTER_FULLNAME
# Arguments:
#   cluster_name
#   cluster_cloud_region
#   cluster_cloud_domain
# Outputs:
#   Writes progress status
#######################################
function aws::init_argocd {
    DEBUG "Deploy ArgoCD via Terraform"
    local cluster_name=$1
    local cluster_cloud_region=$2
    local cluster_cloud_domain=$3

    cd terraform/aws/argocd/ || ERROR "Path not found"

    INFO "ArgoCD: Init Terraform configuration"
    run_cmd "terraform init \
                -backend-config='bucket=$S3_BACKEND_BUCKET' \
                -backend-config='key=$cluster_name/terraform-argocd.state' \
                -backend-config='region=$cluster_cloud_region'"

    run_cmd "terraform plan \
                -var='argo_domain=argo-$CLUSTER_FULLNAME.$cluster_cloud_domain' \
                -input=false \
                -out=tfplan-argocd"

    INFO "ArgoCD: Installing/Reconciling"
    run_cmd "terraform apply -auto-approve -compact-warnings -input=false tfplan-argocd"

    cd - >/dev/null || ERROR "Path not found"
}

#######################################
# Writes commands for user for get access to cluster
# Globals:
#   CLUSTER_FULLNAME
# Arguments:
#   cluster_cloud_domain
# Outputs:
#   Writes commands to get cluster's kubeconfig and ssh key
#######################################
function aws::output_access_keys {
    DEBUG "Writes commands for user for get access to cluster"
    local cluster_cloud_domain=$1

    # TODO: Add output as part of output status. Add commit-back hook with instructions to .cluster.dev/README.md

    KUBECONFIG_DOWNLOAD_MESSAGE="Download and apply your kubeconfig using commands: \n
aws s3 cp s3://${CLUSTER_FULLNAME}/kubeconfig_${CLUSTER_FULLNAME} ~/.kube/kubeconfig_${CLUSTER_FULLNAME} &&
export KUBECONFIG=\$KUBECONFIG:~/.kube/kubeconfig_${CLUSTER_FULLNAME} &&
kubectl get ns
"
    SSH_ACCESS_MESSAGE="Download your bastion ssh key using commands: \n
aws s3 cp s3://${CLUSTER_FULLNAME}/id_rsa_${CLUSTER_FULLNAME}.pem ~/.ssh/id_rsa_${CLUSTER_FULLNAME}.pem && chmod 600 ~/.ssh/id_rsa_${CLUSTER_FULLNAME}.pem &&
ssh -i ~/.ssh/id_rsa_${CLUSTER_FULLNAME}.pem centos@$CLUSTER_FULLNAME.$cluster_cloud_domain
"
    NOTICE "$KUBECONFIG_DOWNLOAD_MESSAGE"
    NOTICE "$SSH_ACCESS_MESSAGE"

    # Add output to GitHub Action Step "steps.reconcile.outputs.(kubeconfig|ssh)"
    echo ::set-output name=kubeconfig::\ "$KUBECONFIG_DOWNLOAD_MESSAGE"
    echo ::set-output name=ssh::\ "$SSH_ACCESS_MESSAGE"

}

#######################################
# Writes information about used software
# Globals:
#   None
# Arguments:
#   None
# Outputs:
#   Writes software versions
#######################################
function main::output_software_info {
    DEBUG "Writes information about used software"
    INFO "Software installed information:"
    INFO "Helm"
    helmfile -v
    INFO "kubectl"
    kubectl version
    INFO "git"
    git --version
    INFO "AWS CLI"
    aws --version
}

#######################################
# Deploy ArgoCD applications via kubectl
# Globals:
#   None
# Arguments:
#   None
# Outputs:
#   Writes progress status
#######################################
function argocd::deploy_apps {
    DEBUG "Deploy ArgoCD apps via kubectl from kubernetes/apps folder"
    DEBUG "Current folder $(pwd)"

    INFO "Deploy Apps into ArgoCD"
    run_cmd "kubectl apply -f kubernetes/apps --recursive"
}
