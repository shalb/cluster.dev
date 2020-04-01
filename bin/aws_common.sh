#!/usr/bin/env bash

# shellcheck source=logging.sh
source "$PRJ_ROOT"/bin/logging.sh

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

    cd "$PRJ_ROOT"/terraform/aws/backend/ || ERROR "Path not found"

    # Check if bucket already exist by trying to import it
    if (aws::is_s3_bucket_exists "$cluster_cloud_region"); then
        INFO "Terraform S3_BACKEND_BUCKET: $S3_BACKEND_BUCKET already exist"
    else
        NOTICE "Terraform S3_BACKEND_BUCKET: $S3_BACKEND_BUCKET not exist. It is going to be created"
        run_cmd "terraform apply -auto-approve \
                    -var='region=$cluster_cloud_region' \
                    -var='s3_backend_bucket=$S3_BACKEND_BUCKET'"
    fi
    run_cmd "rm -rf *.tfstate"

    cd - >/dev/null || ERROR "Path not found"
}

#######################################
# Check if the s3 bucket exists.
#
# Globals:
#   S3_BACKEND_BUCKET
# Arguments:
#   cluster_cloud_region
# Outputs:
#   null
# Return:
#   exitcode
#######################################
function aws::is_s3_bucket_exists {
    local cluster_cloud_region=$1
    cd "$PRJ_ROOT"/terraform/aws/backend/ || ERROR "Path not found"

    # Create and init backend.
    run_cmd "terraform init"

    terraform import -var="region=$cluster_cloud_region" -var="s3_backend_bucket=$S3_BACKEND_BUCKET" aws_s3_bucket.terraform_state "$S3_BACKEND_BUCKET" >/dev/null 2>&1
    return $?
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
    DEBUG "Destroy existing S3 bucket for Terraform states. Bucket name: '${S3_BACKEND_BUCKET}'"
    INFO "Destroying S3 bucket for Terraform states."
    local cluster_cloud_region=$1
    # Delete s3 versions.
    aws s3api delete-objects --bucket "${S3_BACKEND_BUCKET}" --delete "$(aws s3api list-object-versions --bucket ${S3_BACKEND_BUCKET} --output=json --query='{Objects: Versions[].{Key:Key,VersionId:VersionId}}')" > /dev/null 2>&1
    # Delete s3 versions deleted markers.
    aws s3api delete-objects --bucket "${S3_BACKEND_BUCKET}" --delete "$(aws s3api list-object-versions --bucket ${S3_BACKEND_BUCKET} --output=json --query='{Objects: DeleteMarkers[].{Key:Key,VersionId:VersionId}}')" > /dev/null 2>&1
    # Delete bucket.
    run_cmd "aws s3 rb \"s3://${S3_BACKEND_BUCKET}\" --force"
    # Delete dynamodb table.
    aws --region "${cluster_cloud_region}" dynamodb delete-table --table-name "${S3_BACKEND_BUCKET}-state" > /dev/null 2>&1
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

    cd "$PRJ_ROOT"/terraform/aws/route53/ || ERROR "Path not found"

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
    DEBUG "Destroy a DNS domains/records if required"
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
#   FUNC_RESULT
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

    cd "$PRJ_ROOT"/terraform/aws/vpc/ || ERROR "Path not found"

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
    # shellcheck disable=SC2034
    FUNC_RESULT="${cluster_cloud_vpc_id}"

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
    DEBUG "Destroy created VPC keep default unchanged"
    cd "$PRJ_ROOT"/terraform/aws/vpc/ || ERROR "Path not found"

    case ${cluster_cloud_vpc} in
        default|"")
            INFO "Default VPC, no need to destroy."
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
            ;;
    esac

    cd - >/dev/null || ERROR "Path not found"
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

    cd "$PRJ_ROOT"/terraform/aws/argocd/ || ERROR "Path not found"

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

    local output
    output=$(terraform output)
    ARGOCD_ACCESS="ArgoCD Credentials:\n\
${output//$'\n'/'\n'}" # newline characters shielding

    NOTICE "$ARGOCD_ACCESS"

    echo "::set-output name=argocd::${ARGOCD_ACCESS}"

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

    cd "$PRJ_ROOT"/terraform/aws/argocd/ || ERROR "Path not found"

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

    KUBECONFIG_DOWNLOAD_MESSAGE="\
Download and apply your kubeconfig using commands: \n\
aws s3 cp s3://${CLUSTER_FULLNAME}/kubeconfig_${CLUSTER_FULLNAME} ~/.kube/kubeconfig_${CLUSTER_FULLNAME} \n\
export KUBECONFIG=\$KUBECONFIG:~/.kube/kubeconfig_${CLUSTER_FULLNAME} \n\
kubectl get ns \n
"
    SSH_ACCESS_MESSAGE="\
Download your bastion ssh key using commands: \n\
aws s3 cp s3://${CLUSTER_FULLNAME}/id_rsa_${CLUSTER_FULLNAME}.pem ~/.ssh/id_rsa_${CLUSTER_FULLNAME}.pem && chmod 600 ~/.ssh/id_rsa_${CLUSTER_FULLNAME}.pem \n\
ssh -i ~/.ssh/id_rsa_${CLUSTER_FULLNAME}.pem ubuntu@$CLUSTER_FULLNAME.$cluster_cloud_domain \n
"

    NOTICE "$KUBECONFIG_DOWNLOAD_MESSAGE"
    NOTICE "$SSH_ACCESS_MESSAGE"

    # Add output to GitHub Action Step "steps.reconcile.outputs.(kubeconfig|ssh)"
    echo "::set-output name=kubeconfig::${KUBECONFIG_DOWNLOAD_MESSAGE}"
    echo "::set-output name=ssh::${SSH_ACCESS_MESSAGE}"

}


# Destroy all cluster.
function aws::destroy {
        case $cluster_provisioner_type in
        minikube)
            DEBUG "Destroy: Provisioner: Minikube"
            if aws::minikube::pull_kubeconfig_once; then
                # aws::destroy_argocd "$cluster_name" "$cluster_cloud_region" "$cluster_cloud_domain"
                kube::destroy_apps
            fi
            aws::minikube::destroy_cluster "$cluster_name" "$cluster_cloud_region" "$cluster_provisioner_instanceType" "$cluster_cloud_domain"
            # TODO: Remove kubeconfig after successful cluster destroy
            aws::destroy_vpc "$cluster_cloud_vpc" "$cluster_name" "$cluster_cloud_region"
            aws::destroy_route53 "$cluster_cloud_region" "$cluster_name" "$cluster_cloud_domain"
            aws::destroy_s3_bucket "$cluster_cloud_region"
        ;;
        # end of minikube
        eks)
            DEBUG "Destroy: Provisioner: EKS"
            ;;
        esac
}
