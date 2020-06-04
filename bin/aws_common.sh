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
#   cluster_cloud_domain
#######################################
function aws::init_route53 {
    DEBUG "Create a DNS domains/records if required"
    local default_domain="cluster.dev"
    local cluster_cloud_region=$1
    local cluster_name=$2
    local cluster_cloud_domain=${3:-$default_domain}

    # Init terraform state for DNS
    cd "$PRJ_ROOT"/terraform/aws/route53/ || ERROR "Path not found"
    terraform init -backend-config="bucket=$S3_BACKEND_BUCKET" \
        -backend-config="key=states/terraform-dns.state" \
        -backend-config="region=$cluster_cloud_region"

    # Create or update zone
    if [ "$cluster_cloud_domain" = "$default_domain" ]; then
        INFO "The cluster domain is unset. DNS sub-zone would be created in $default_domain"
        zone_delegation=true
    else
        INFO "The cluster domain defined. DNS sub-zone would be created in $cluster_cloud_domain"
        zone_delegation=false
    fi
    # Execute terraform
    run_cmd "terraform plan -compact-warnings \
            -var='region=$cluster_cloud_region' \
            -var='cluster_name=$cluster_name' \
            -var='cluster_domain=$cluster_cloud_domain' \
            -var='zone_delegation=$zone_delegation' \
            -input=false \
            -out=tfplan"
    run_cmd "terraform apply -auto-approve -compact-warnings -input=false tfplan"
    INFO "DNS Zone: $cluster_name.$cluster_cloud_domain has been created."

    cd - >/dev/null || ERROR "Path not found"
}

#######################################
# Destroy a DNS domains/records if required
# Globals:
#   S3_BACKEND_BUCKET
#   CLUSTER_FULLNAME
# Arguments:
#   cluster_cloud_region
#   cluster_cloud_domain
#######################################
function aws::destroy_route53 {
    local default_domain="cluster.dev"
    local cluster_cloud_region=$1
    local cluster_name=$2
    local cluster_cloud_domain=${3:-$default_domain}

    # Init terraform state for DNS
    cd "$PRJ_ROOT"/terraform/aws/route53/ || ERROR "Path not found"
    terraform init -backend-config="bucket=$S3_BACKEND_BUCKET" \
        -backend-config="key=states/terraform-dns.state" \
        -backend-config="region=$cluster_cloud_region"

    # Create or update zone
    if [ "$cluster_cloud_domain" = "$default_domain" ]; then
        INFO "The cluster domain is unset. DNS sub-zone would be created in $default_domain"
        zone_delegation=true
    else
        INFO "The cluster domain defined. DNS sub-zone would be created in $cluster_cloud_domain"
        zone_delegation=false
    fi

    # Execute terraform
    INFO "Destroying a DNS zone $cluster_name.$cluster_cloud_domain"
    run_cmd "terraform  destroy -auto-approve  \
            -var='region=$cluster_cloud_region' \
            -var='cluster_domain=$cluster_cloud_domain' \
            -var='zone_delegation=$zone_delegation' \
            -var='cluster_name=$cluster_name'"

    INFO "DNS Zone: $cluster_name.$cluster_cloud_domain has been deleted."

    cd - >/dev/null || ERROR "Path not found"
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
#   cluster_cloud_availability_zones
#   cluster_cloud_vpc_cidr
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
    local availability_zones=$4
    availability_zones=$(to_tf_list "$availability_zones") # convert to terraform list format
    local vpc_cidr=${5:-"10.8.0.0/18"} # set default VPC cidr

    cd "$PRJ_ROOT"/terraform/aws/vpc/ || ERROR "Path not found"

            # Create/Init VPC and get ID.
            INFO "VPC: Initializing Terraform configuration"
            run_cmd "terraform init \
                        -backend-config='bucket=$S3_BACKEND_BUCKET' \
                        -backend-config='key=states/terraform-vpc.state' \
                        -backend-config='region=$cluster_cloud_region'"

            run_cmd "terraform plan \
                        -var='vpc_id=$cluster_cloud_vpc' \
                        -var='cluster_name=$CLUSTER_FULLNAME' \
                        -var='region=$cluster_cloud_region' \
                        -var='availability_zones=$availability_zones' \
                        -var='vpc_cidr=$vpc_cidr' \
                        -input=false \
                        -out=tfplan"

            INFO "VPC: Apply infrastructure changes"
            run_cmd "terraform apply -auto-approve -compact-warnings -input=false tfplan"

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
#   availability_zones
# Outputs:
#   Writes progress status
#######################################
function aws::destroy_vpc {
    local cluster_cloud_vpc=$1
    local availability_zones=$4
    availability_zones=$(to_tf_list "$availability_zones") # convert to terraform list format

    DEBUG "Destroy created VPC keep default unchanged"
    cd "$PRJ_ROOT"/terraform/aws/vpc/ || ERROR "Path not found"

            INFO "VPC: Initializing Terraform configuration"
            run_cmd "terraform init \
                        -backend-config='bucket=$S3_BACKEND_BUCKET' \
                        -backend-config='key=states/terraform-vpc.state' \
                        -backend-config='region=$cluster_cloud_region'"

            INFO "VPC: Destroying"
            run_cmd "terraform destroy -auto-approve -compact-warnings \
                        -var='region=$cluster_cloud_region' \
                        -var='availability_zones=$availability_zones' \
                        -var='vpc_id=$cluster_cloud_vpc' \
                        -var='cluster_name=$CLUSTER_FULLNAME'"

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
export KUBECONFIG=~/.kube/kubeconfig_${CLUSTER_FULLNAME} \n\
kubectl get ns \n
"
    NOTICE "$KUBECONFIG_DOWNLOAD_MESSAGE"

#    SSH_ACCESS_MESSAGE="\
#Download your bastion ssh key using commands: \n\
#aws s3 cp s3://${CLUSTER_FULLNAME}/id_rsa_${CLUSTER_FULLNAME}.pem ~/.ssh/id_rsa_${CLUSTER_FULLNAME}.pem && chmod 600 ~/.ssh/id_rsa_${CLUSTER_FULLNAME}.pem \n\
#ssh -i ~/.ssh/id_rsa_${CLUSTER_FULLNAME}.pem ubuntu@$CLUSTER_FULLNAME.$cluster_cloud_domain \n
#"
#    NOTICE "$SSH_ACCESS_MESSAGE"


    # Add output to GitHub Action Step "steps.reconcile.outputs.(kubeconfig|ssh)"
    echo "::set-output name=kubeconfig::${KUBECONFIG_DOWNLOAD_MESSAGE}"
#    echo "::set-output name=ssh::${SSH_ACCESS_MESSAGE}"

}

# Destroy all cluster.
function aws::destroy {

        case $cluster_cloud_provisioner_type in
        minikube)
            DEBUG "Destroy: Provisioner: Minikube"
            if aws::minikube::pull_kubeconfig_once; then
                aws::destroy_addons "$CLUSTER_FULLNAME" "$cluster_cloud_region" "$cluster_cloud_domain"
            fi
            aws::minikube::destroy_cluster "$CLUSTER_FULLNAME" "$cluster_cloud_region" "$cluster_cloud_domain" "$cluster_cloud_provisioner_instanceType"
        ;;
        # end of minikube
        eks)
            DEBUG "Destroy: Provisioner: EKS"
            # Destroy Cluster
            aws::eks::destroy_cluster "$CLUSTER_FULLNAME" "$cluster_cloud_region" "$cluster_cloud_availability_zones" "$cluster_cloud_domain"
            ;;
        esac

        # Destroy all cluster components
        # TODO: Remove kubeconfig after successful cluster destroy
        aws::destroy_vpc "$cluster_cloud_vpc" "$CLUSTER_FULLNAME" "$cluster_cloud_region" "$cluster_cloud_availability_zones"
        aws::destroy_route53 "$cluster_cloud_region" "$CLUSTER_FULLNAME" "$cluster_cloud_domain"
        aws::destroy_s3_bucket "$cluster_cloud_region"
}

#######################################
# Deploy K8s Addons via Terraform
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
function aws::init_addons {
    DEBUG "Deploy Kubernetes Addons via Terraform"
    local cluster_name=$1
    local cluster_cloud_region=$2
    local cluster_cloud_domain=$3
    local config_path=${4:-"~/.kube/config"}

    cd "$PRJ_ROOT"/terraform/aws/addons/ || ERROR "Path not found"

    INFO "Kubernetes Addons: Init Terraform configuration"
    run_cmd "terraform init \
                -backend-config='bucket=$S3_BACKEND_BUCKET' \
                -backend-config='key=states/terraform-addons.state' \
                -backend-config='region=$cluster_cloud_region'"

    run_cmd "terraform plan \
                -var='region=$cluster_cloud_region' \
                -var='cluster_cloud_domain=$cluster_cloud_domain' \
                -var='cluster_name=$CLUSTER_FULLNAME' \
                -var='config_path=$config_path' \
                -var='eks=true' \
                -input=false \
                -out=tfplan-addons"

    INFO "Kubernetes Addons: Installing/Reconciling"
    run_cmd "terraform apply -auto-approve -compact-warnings -input=false tfplan-addons"

    local output
    output=$(terraform output)
    ARGOCD_ACCESS="ArgoCD Credentials:\n\
${output//$'\n'/'\n'}" # newline characters shielding

    NOTICE "$ARGOCD_ACCESS"

    echo "::set-output name=argocd::${ARGOCD_ACCESS}"

    cd - >/dev/null || ERROR "Path not found"
}

#######################################
# Destroy Kubernetes Addons via Terraform
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
function aws::destroy_addons {
    DEBUG "Delete Kubernetes Addons via Terraform"
    local cluster_name=$1
    local cluster_cloud_region=$2
    local cluster_cloud_domain=$3

    cd "$PRJ_ROOT"/terraform/aws/addons/ || ERROR "Path not found"

    INFO "Kubernetes Addons: Init Terraform configuration"
    run_cmd "terraform init \
                -backend-config='bucket=$S3_BACKEND_BUCKET' \
                -backend-config='key=states/terraform-addons.state' \
                -backend-config='region=$cluster_cloud_region'"

    INFO "Kubernetes Addons: Destroying"
    run_cmd "terraform destroy -auto-approve -compact-warnings \
                -var='cluster_cloud_domain=$cluster_cloud_domain' \
                -var='region=$cluster_cloud_region'" "" "false"

    cd - >/dev/null || ERROR "Path not found"
}
