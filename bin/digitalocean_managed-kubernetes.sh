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
#   cluster_cloud_provisioner_nodeCount
# Outputs:
#   Writes progress status
#######################################
function digitalocean::managed-kubernetes::deploy_cluster {
    DEBUG "Deploy DO Kubernetes cluster via Terraform"
    local cluster_name=$1
    local cluster_cloud_region=$2
    local cluster_cloud_provisioner_version=$3
    local cluster_cloud_provisioner_nodeSize=$4
    local cluster_cloud_provisioner_nodeCount=$5

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
                  -var='node_count=$cluster_cloud_provisioner_nodeCount' \
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
#   cluster_cloud_provisioner_nodeCount
# Outputs:
#   Writes progress status
#######################################
function digitalocean::managed-kubernetes::destroy_cluster {
    DEBUG "Destroy DO Kubernetes cluster via Terraform"
    local cluster_name=$1
    local cluster_cloud_region=$2
    local cluster_cloud_provisioner_version=$3
    local cluster_cloud_provisioner_nodeSize=$4
    local cluster_cloud_provisioner_nodeCount=$5

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
                -var='k8s_version=$cluster_cloud_provisioner_version' \
                -var='name=$CLUSTER_FULLNAME' \
                -var='node_type=$cluster_cloud_provisioner_nodeSize' \
                -var='node_count=$cluster_cloud_provisioner_nodeCount'"

    cd - >/dev/null || ERROR "Path not found"
}

#######################################
# Deploy DO Kubernetes cluster with autoscaling enabled feature via Terraform
# Globals:
#   DO_SPACES_BACKEND_BUCKET
#   CLUSTER_FULLNAME
# Arguments:
#   cluster_name
#   cluster_cloud_region
#   cluster_cloud_provisioner_version
#   cluster_cloud_provisioner_nodeSize
#   cluster_cloud_provisioner_autoScale
#   cluster_cloud_provisioner_minNodes
#   cluster_cloud_provisioner_maxNodes
# Outputs:
#   Writes progress status
#######################################
function digitalocean::managed-kubernetes::deploy_cluster_autoscale {
    DEBUG "Deploy DO Kubernetes cluster with autoscaling enabled feature via Terraform"
    local cluster_name=$1
    local cluster_cloud_region=$2
    local cluster_cloud_provisioner_version=$3
    local cluster_cloud_provisioner_nodeSize=$4
    local cluster_cloud_provisioner_autoScale=$5
    local cluster_cloud_provisioner_minNodes=$6
    local cluster_cloud_provisioner_maxNodes=$7

    cd "$PRJ_ROOT"/terraform/digitalocean/k8s/ || ERROR "Path not found"

    # Deploy main Terraform code
    INFO "DO k8s cluster with autoscaling enabled feature: Initializing Terraform configuration"
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
                  -var='enable_autoscaling=$cluster_cloud_provisioner_autoScale' \
                  -var='min_node_count=$cluster_cloud_provisioner_minNodes' \
                  -var='max_node_count=$cluster_cloud_provisioner_maxNodes' \
                  -input=false \
                  -out=tfplan"

    INFO "DO k8s cluster with autoscaling feature: Creating infrastructure"
    run_cmd "terraform apply -auto-approve -compact-warnings -input=false tfplan"

    cd - >/dev/null || ERROR "Path not found"
}

#######################################
# Destroy DO Kubernetes cluster with autoscaling enabled feature via Terraform
# Globals:
#   DO_SPACES_BACKEND_BUCKET
#   CLUSTER_FULLNAME
# Arguments:
#   cluster_name
#   cluster_cloud_region
#   cluster_cloud_provisioner_version
#   cluster_cloud_provisioner_nodeSize
#   cluster_cloud_provisioner_autoScale
#   cluster_cloud_provisioner_minNodes
#   cluster_cloud_provisioner_maxNodes
# Outputs:
#   Writes progress status
#######################################
function digitalocean::managed-kubernetes::destroy_cluster_autoscale {
    DEBUG "Destroy DO Kubernetes cluster with autoscaling enabled feature via Terraform"
    local cluster_name=$1
    local cluster_cloud_region=$2
    local cluster_cloud_provisioner_version=$3
    local cluster_cloud_provisioner_nodeSize=$4
    local cluster_cloud_provisioner_autoScale=$5
    local cluster_cloud_provisioner_minNodes=$6
    local cluster_cloud_provisioner_maxNodes=$7

    cd "$PRJ_ROOT"/terraform/digitalocean/k8s/ || ERROR "Path not found"

    # Deploy main Terraform code
    INFO "DO k8s cluster with autoscaling enabled feature: Initializing Terraform configuration"
    run_cmd "terraform init \
                -backend-config='bucket=$DO_SPACES_BACKEND_BUCKET' \
                -backend-config='key=$cluster_name/terraform.state' \
                -backend-config='endpoint=$cluster_cloud_region.digitaloceanspaces.com' \
                -backend-config='access_key=$SPACES_ACCESS_KEY_ID' \
                -backend-config='secret_key=$SPACES_SECRET_ACCESS_KEY'"

    INFO "DO k8s cluster with autoscaling enabled feature: Destroying "
    run_cmd "terraform destroy -auto-approve -compact-warnings \
                -var='region=$cluster_cloud_region' \
                -var='k8s_version=$cluster_cloud_provisioner_version' \
                -var='name=$CLUSTER_FULLNAME' \
                -var='node_type=$cluster_cloud_provisioner_nodeSize' \
                -var='enable_autoscaling=$cluster_cloud_provisioner_autoScale' \
                -var='min_node_count=$cluster_cloud_provisioner_minNodes' \
                -var='max_node_count=$cluster_cloud_provisioner_maxNodes'"

    cd - >/dev/null || ERROR "Path not found"
}
