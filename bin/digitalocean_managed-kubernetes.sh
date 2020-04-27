#!/usr/bin/env bash

# shellcheck source=logging.sh
source "$PRJ_ROOT"/bin/logging.sh

#######################################
# Deploy DO Kubernetes cluster via Terraform
# Globals:
#   DO_SPACES_BACKEND_BUCKET
#   CLUSTER_FULLNAME
# Arguments:
#   cluster_name
#   cluster_cloud_region
#   cluster_cloud_provisioner_version
#   cluster_cloud_provisioner_nodeSize
# Outputs:
#   Writes progress status
#######################################
function digitalocean::managed-kubernetes::deploy_cluster {
    DEBUG "Deploy DO Kubernetes cluster via Terraform"
    local cluster_name=$1
    local cluster_cloud_region=$2
    local cluster_cloud_provisioner_version=$3
    local cluster_cloud_provisioner_nodeSize=$4

    cd "$PRJ_ROOT"/terraform/digitalocean/k8s/ || ERROR "Path not found"

    # Deploy main Terraform code
    INFO "DO k8s cluster: Initializing Terraform configuration"
    run_cmd "terraform init \
                -backend-config='bucket=$DO_SPACES_BACKEND_BUCKET' \
                -backend-config='key=$cluster_name/terraform.state' \
                -backend-config='endpoint=$cluster_cloud_region.digitaloceanspaces.com' \
                -backend-config='access_key=$SPACES_ACCESS_KEY_ID' \
                -backend-config='secret_key=$SPACES_SECRET_ACCESS_KEY'"


    run_cmd "terraform plan \
                -var='region=$cluster_cloud_region' \
                -var='k8s_version=$cluster_cloud_provisioner_version' \
                -var='name=$CLUSTER_FULLNAME' \
                -var='node_type=$cluster_cloud_provisioner_nodeSize' \
                -input=false \
                -out=tfplan"

    INFO "DO k8s cluster: Creating infrastructure"
    run_cmd "terraform apply -auto-approve -compact-warnings -input=false tfplan"

    cd - >/dev/null || ERROR "Path not found"
}

#######################################
# Destroy DO Kubernetes cluster via Terraform
# Globals:
#   DO_SPACES_BACKEND_BUCKET
#   CLUSTER_FULLNAME
# Arguments:
#   cluster_name
#   cluster_cloud_region
#   cluster_cloud_provisioner_version
#   cluster_cloud_provisioner_nodeSize
# Outputs:
#   Writes progress status
#######################################
function digitalocean::managed-kubernetes::destroy_cluster {
    DEBUG "Destroy DO Kubernetes cluster via Terraform"
    local cluster_name=$1
    local cluster_cloud_region=$2
    local cluster_cloud_provisioner_version=$3
    local cluster_cloud_provisioner_nodeSize=$4

    cd "$PRJ_ROOT"/terraform/digitalocean/k8s/ || ERROR "Path not found"

    # Deploy main Terraform code
    INFO "DO k8s cluster: Initializing Terraform configuration"
    run_cmd "terraform init \
                -backend-config='bucket=$DO_SPACES_BACKEND_BUCKET' \
                -backend-config='key=$cluster_name/terraform.state' \
                -backend-config='endpoint=$cluster_cloud_region.digitaloceanspaces.com' \
                -backend-config='access_key=$SPACES_ACCESS_KEY_ID' \
                -backend-config='secret_key=$SPACES_SECRET_ACCESS_KEY'"


    INFO "DO k8s cluster: Destroying "
    run_cmd "terraform destroy -auto-approve -compact-warnings \
                -var='region=$cluster_cloud_region' \
                -var='k8_version=$cluster_cloud_provisioner_version' \
                -var='name=$CLUSTER_FULLNAME' \
                -var='node_type=$cluster_cloud_provisioner_nodeSize'"

    cd - >/dev/null || ERROR "Path not found"
}
