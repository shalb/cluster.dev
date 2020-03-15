#!/usr/bin/env bash

# Parse YAML configs in .cluster-dev/*
source ./bin/yaml.sh # provides parse_yaml and create_variables
source ./bin/bash-logger.sh # PSR-3 compliant logging

# Variables passed by Github Workflow to Action
readonly CLUSTER_CONFIG_PATH=$1
readonly CLOUD_USER=$2
readonly CLOUD_PASS=$3
# For local testing run: ./entrypoint.sh .cluster.dev/minikube-one.yaml AWSUSER AWSPASS
#

#######################################
# Run command into wrapper, that print command output only when:
# - error happened
# - VERBOSE_LVL=DEBUG
# Globals:
#   None
# Arguments:
#   command - command that should be executed inside wrapper
#   bash_opts - additional shell options
#   fail_on_err - interpret not 0 exit code as error or not.
#                 Boolean. By default - true.
#   enable_log_timeout - Print all logs for command after timeout.
#                        Useful for non-DEBUG log levels.
#                        By default - 300 seconds (5 min)
# Outputs:
#   Writes progress status
#######################################
function run_cmd {
    local command="$1"
    local bash_opts="${2-""}"
    local fail_on_err="${3-true}"
    local enable_log_timeout="${3-300}" # By default - 300 seconds (5 min)

    local bash="/usr/bin/env bash ${bash_opts}"

    # STDERR prints by default
    # Log STDOUT and continue
    if [ "$LOG_LVL" == "DEBUG" ]; then
        DEBUG "Output from '$command'" "" "" 1
        ${bash} -x -c "$command"
        exit_code=$?

        [ ${exit_code} != 0 ] && [[ $fail_on_err ]] && ERROR "Execution of '$command' failed" "" "" 1
        return
    fi

    # shellcheck disable=SC2091
    $(${bash} -x -c "$command" >/tmp/log) &
    proc_pid=$!

    # Print logs in non-debug
    seconds=0
    until [ ! -d /proc/$proc_pid ]; do
        if [[ $seconds -gt $enable_log_timeout ]]; then
            WARNING "Command '$command' executing more than ${seconds}s. \
Start print out exist and future output for this command." "%DATE - %MESSAGE" "+%T" 1
            tail -f -n +0 /tmp/log &
            tail_pid=$!
            break
        fi
        sleep 1
        ((seconds++))
    done

    wait $proc_pid
    exit_code=$?
    # `> /dev/null 2>&1 &` and `wait` need to suppress "Terminated' message
    kill $tail_pid > /dev/null 2>&1 &
    wait

    [ ${exit_code} != 0 ] && [[ $fail_on_err ]] && ERROR "Execution of '$command' failed" "" "" 1
}

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
# Destroy S3 bucket for Terraform states
# Globals:
#   S3_BACKEND_BUCKET
# Arguments:
#   cluster_cloud_region
# Outputs:
#   Writes progress status
#######################################
function aws::destroy_s3_bucket {
    DEBUG "Create or use exiting S3 bucket for Terraform states"
    local cluster_cloud_region=$1

    #TODO: remove bucket procedure
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
# Destroy a DNS domains/records if required
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
function aws::destroy_route53 {
    DEBUG "Create a DNS domains/records if required"
    local cluster_cloud_region=$1
    local cluster_name=$2
    local cluster_cloud_domain=$3


    # TODO: destroy procedure.
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
# Destroy a VPC or use existing defined
# Globals:
#   S3_BACKEND_BUCKET
#   CLUSTER_FULLNAME
# Arguments:
#   cluster_cloud_vpc
#   cluster_name
#   cluster_cloud_region
# Outputs:
#   Writes progress status
#######################################
function aws::destroy_vpc {
    local cluster_cloud_vpc=$1

    cd terraform/aws/vpc/ || ERROR "Path not found"

    case ${cluster_cloud_vpc} in
        default|"")
            INFO "Default VPC, no need to destroy."
            return
            ;;
        create)
            # Create new VPC and get ID.
            INFO "VPC: Initializing Terraform configuration"
            run_cmd "terraform init \
                        -backend-config='bucket=$S3_BACKEND_BUCKET' \
                        -backend-config='key=$cluster_name/terraform-vpc.state' \
                        -backend-config='region=$cluster_cloud_region'"

            INFO "VPC: Destroying"
            run_cmd "terraform destroy -auto-approve -compact-warnings \
                        -var='region=$cluster_cloud_region' \
                        -var='cluster_name=$CLUSTER_FULLNAME'"
            ;;
        *)
            # Use client VPC ID.
            INFO "Custom VPC, no need to destroy."
            return
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
# Try get kubeconfig to instance via kubectl
# Globals:
#   CLUSTER_FULLNAME
# Arguments:
#   None
# Outputs:
#   Writes progress status
#######################################
function aws::minikube::pull_kubeconfig_try {
    DEBUG "Pull a kubeconfig to instance via kubectl"
    local WAIT_TIMEOUT=5

    INFO "Copy kubeconfig to instance with Minikube"
    export KUBECONFIG=~/.kube/kubeconfig_${CLUSTER_FULLNAME}
    run_cmd "aws s3 cp 's3://${CLUSTER_FULLNAME}/kubeconfig_$CLUSTER_FULLNAME' '$HOME/.kube/kubeconfig_$CLUSTER_FULLNAME' 2>/dev/null"
    run_cmd "cp '$HOME/.kube/kubeconfig_$CLUSTER_FULLNAME' '$HOME/.kube/config' 2>/dev/null"
    kubectl version --request-timeout=5s >/dev/null 2>&1
    return $?
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
# Destroy Minikube cluster via Terraform
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
function aws::minikube::destroy_cluster {
    DEBUG "Destroy Minikube cluster via Terraform"
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

    INFO "Minikube cluster: Destroying "
    run_cmd "terraform destroy -auto-approve -compact-warnings \
                -var='region=$cluster_cloud_region' \
                -var='cluster_name=$CLUSTER_FULLNAME' \
                -var='aws_instance_type=$cluster_provisioner_instanceType' \
                -var='hosted_zone=$cluster_cloud_domain'"

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
function kube::deploy_cert_manager {
    DEBUG "Deploy CertManager via kubectl"

    INFO "Setup TLS certificates"
    run_cmd "kubectl apply -f 'https://raw.githubusercontent.com/shalb/terraform-aws-minikube/master/addons/ingress.yaml'"
    run_cmd "kubectl apply -f 'https://github.com/jetstack/cert-manager/releases/download/v0.13.0/cert-manager-no-webhook.yaml'"
    run_cmd "kubectl apply -f 'https://raw.githubusercontent.com/shalb/terraform-aws-minikube/8a147f7c0044c318ec37990b50f0cabb205e9b44/addons/letsencrypt-prod.yaml'"
}

#######################################
# Remove CertManager via kubectl
# Globals:
#   None
# Arguments:
#   None
# Outputs:
#   Writes progress status
#######################################
function kube::destroy_cert_manager {
    DEBUG "Deploy CertManager via kubectl"

    INFO "Setup TLS certificates"
    run_cmd "kubectl delete -f 'https://raw.githubusercontent.com/shalb/terraform-aws-minikube/8a147f7c0044c318ec37990b50f0cabb205e9b44/addons/letsencrypt-prod.yaml' || true"
    run_cmd "kubectl delete -f 'https://github.com/jetstack/cert-manager/releases/download/v0.13.0/cert-manager-no-webhook.yaml' || true"
    run_cmd "kubectl delete -f 'https://raw.githubusercontent.com/shalb/terraform-aws-minikube/master/addons/ingress.yaml' || true"
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
# Destroy ArgoCD via Terraform
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
function aws::destroy_argocd {
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

    INFO "ArgoCD: Destroying"
    run_cmd "terraform destroy -auto-approve -compact-warnings \
                -var='argo_domain=argo-$CLUSTER_FULLNAME.$cluster_cloud_domain'"

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
    NOTICE $KUBECONFIG_DOWNLOAD_MESSAGE
    NOTICE $SSH_ACCESS_MESSAGE

# Add output to GitHub Action Step "steps.reconcile.outputs.(kubeconfig|ssh)"
echo ::set-output name=kubeconfig::\ $KUBECONFIG_DOWNLOAD_MESSAGE
echo ::set-output name=ssh::\ $SSH_ACCESS_MESSAGE

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

# Destroy all cluster.
function aws::destroy {

        case $cluster_provisioner_type in
        minikube)
            DEBUG "Destroy: Provisioner: Minikube"
            if aws::minikube::pull_kubeconfig_try; then
                aws::destroy_argocd "$cluster_name" "$cluster_cloud_region" "$cluster_cloud_domain"
                kube::destroy_cert_manager
            fi
            aws::minikube::destroy_cluster "$cluster_name" "$cluster_cloud_region" "$cluster_provisioner_instanceType" "$cluster_cloud_domain"
            aws::destroy_vpc "$cluster_cloud_vpc" "$cluster_name" "$cluster_cloud_region"
            aws::destroy_route53 "$cluster_cloud_region" "$cluster_name" "$cluster_cloud_domain"
            aws::destroy_s3_bucket "$cluster_cloud_region"
        ;;
        # end of minikube
        eks)
            DEBUG "Cloud Provider: AWS. Provisioner: EKS"
            ;;
        esac
}


# =========================================================================== #
#                                    MAIN                                     #
# =========================================================================== #


DEBUG "Starting job in repo: $GITHUB_REPOSITORY with arguments  \
      CLUSTER_CONFIG_PATH: $CLUSTER_CONFIG_PATH, CLOUD_USER: $CLOUD_USER"

# Writes information about used software
main::output_software_info

# Iterate trough provided manifests and reconcile clusters
for CLUSTER_MANIFEST_FILE in $(find "$CLUSTER_CONFIG_PATH" -type f); do

    parse_yaml "$CLUSTER_MANIFEST_FILE"
    create_variables "$CLUSTER_MANIFEST_FILE"

    # Cloud selection
    case $cluster_cloud_provider in
    aws)

        DEBUG "Cloud Provider: AWS. Initializing access variables"
        # Define AWS credentials
        export AWS_ACCESS_KEY_ID=$CLOUD_USER
        export AWS_SECRET_ACCESS_KEY=$CLOUD_PASS
        export AWS_DEFAULT_REGION=$cluster_cloud_region
        export CLUSTER_PREFIX=$GITHUB_REPOSITORY # CLUSTER_PREFIX equals git organization/username could be changed in other repo

        # create unique s3 bucket from repo name and cluster name
        S3_BACKEND_BUCKET=$(echo "$CLUSTER_PREFIX" | awk -F "/" '{print$1}')-$cluster_name
        # make sure it is not larger than 63 symbols
        S3_BACKEND_BUCKET=$(echo "$S3_BACKEND_BUCKET" | cut -c 1-63)
        # The same name would be used for domains
        readonly CLUSTER_FULLNAME=$S3_BACKEND_BUCKET

        # Destroy if installed: false
        if [ "$cluster_installed" = "false" ]; then
            aws::destroy
            exit 0

        fi

        # Create and init backend.
        # Check if bucket already exist by trying to import it
        aws::init_s3_bucket   "$cluster_cloud_region"

        # Create a DNS domains/records if required
        aws::init_route53   "$cluster_cloud_region" "$cluster_name" "$cluster_cloud_domain"

        # Create a VPC or use existing defined
        aws::init_vpc   "$cluster_cloud_vpc" "$cluster_name" "$cluster_cloud_region"

        # Provisioner selection
        case $cluster_provisioner_type in
        minikube)
            DEBUG "Provisioner: Minikube"

            # Deploy Minikube cluster via Terraform
            aws::minikube::deploy_cluster   "$cluster_name" "$cluster_cloud_region" "$cluster_provisioner_instanceType" "$cluster_cloud_domain"

            # Pull a kubeconfig to instance via kubectl
            aws::minikube::pull_kubeconfig

            # Deploy CertManager via kubectl
            kube::deploy_cert_manager

            # Deploy ArgoCD via Terraform
            aws::init_argocd   "$cluster_name" "$cluster_cloud_region" "$cluster_cloud_domain"

            # Writes commands for user for get access to cluster
            aws::output_access_keys   "$cluster_cloud_domain"
        ;;
        # end of minikube
        eks)
            DEBUG "Cloud Provider: AWS. Provisioner: EKS"
            ;;
        esac
        ;;

    gcp)
        DEBUG "Cloud Provider: Google"
        ;;

    azure)
        DEBUG "Cloud Provider: Azure"
        ;;

    digitalocean)
        DEBUG "Cloud Provider: Azure"
        ;;

    esac

done
