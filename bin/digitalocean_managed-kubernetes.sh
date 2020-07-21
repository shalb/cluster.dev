#!/usr/bin/env bash

# shellcheck source=logging.sh
source "$PRJ_ROOT"/bin/logging.sh

#######################################
# Deploy DO Kubernetes cluster with autoscaling enabled feature via Terraform
# Globals:
#   DO_SPACES_BACKEND_BUCKET
# Arguments:
#   cluster_name
#   cluster_cloud_region
#   cluster_cloud_provisioner_version
#   cluster_cloud_provisioner_nodeSize
#   cluster_cloud_provisioner_minNodes
#   cluster_cloud_provisioner_maxNodes
# Outputs:
#   Writes progress status
#######################################
function digitalocean::managed-kubernetes::deploy_cluster {
    DEBUG "Deploy DO Kubernetes cluster with autoscaling enabled feature via Terraform"
    local cluster_name=$1
    local cluster_cloud_region=$2
    local cluster_cloud_provisioner_version=$3
    local cluster_cloud_provisioner_nodeSize=$4
    local cluster_cloud_provisioner_minNodes=$5
    local cluster_cloud_provisioner_maxNodes=$6

    cd "$PRJ_ROOT"/terraform/digitalocean/k8s/ || ERROR "Path not found"

    # Deploy main Terraform code
    INFO "DO Managed Kubernetes Cluster: Initializing Terraform configuration"
    run_cmd "terraform init \
                -backend-config='bucket=$DO_SPACES_BACKEND_BUCKET' \
                -backend-config='key=states/terraform-k8s.state' \
                -backend-config='endpoint=$cluster_cloud_region.digitaloceanspaces.com'"

    run_cmd "terraform plan \
                -var='region=$cluster_cloud_region' \
                -var='k8s_version=$cluster_cloud_provisioner_version' \
                -var='cluster_name=$cluster_name' \
                -var='node_type=$cluster_cloud_provisioner_nodeSize' \
                -var='min_node_count=$cluster_cloud_provisioner_minNodes' \
                -var='max_node_count=$cluster_cloud_provisioner_maxNodes' \
                -input=false \
                -out=tfplan"

    INFO "DO Managed Kubernetes Cluster: Creating..."
    run_cmd "terraform apply -auto-approve -compact-warnings -input=false tfplan"

    cd - >/dev/null || ERROR "Path not found"
}

#######################################
# Destroy DO Kubernetes cluster with autoscaling enabled feature via Terraform
# Globals:
#   DO_SPACES_BACKEND_BUCKET
# Arguments:
#   cluster_name
#   cluster_cloud_region
#   cluster_cloud_provisioner_version
#   cluster_cloud_provisioner_nodeSize
#   cluster_cloud_provisioner_minNodes
#   cluster_cloud_provisioner_maxNodes
# Outputs:
#   Writes progress status
#######################################
function digitalocean::managed-kubernetes::destroy_cluster {
    DEBUG "Destroy DO Kubernetes cluster with autoscaling enabled feature via Terraform"
    local cluster_name=$1
    local cluster_cloud_region=$2
    local cluster_cloud_provisioner_version=$3

    cd "$PRJ_ROOT"/terraform/digitalocean/k8s/ || ERROR "Path not found"

    # Deploy main Terraform code
    INFO "DO Managed Kubernetes Cluster: Initializing Terraform configuration"
    run_cmd "terraform init \
                -backend-config='bucket=$DO_SPACES_BACKEND_BUCKET' \
                -backend-config='key=states/terraform-k8s.state' \
                -backend-config='endpoint=$cluster_cloud_region.digitaloceanspaces.com'"

    INFO "DO Managed Kubernetes Cluster: Destroying"
    run_cmd "terraform destroy -auto-approve -compact-warnings \
                -var='region=$cluster_cloud_region' \
                -var='k8s_version=$cluster_cloud_provisioner_version' \
                -var='cluster_name=$cluster_name'"

    cd - >/dev/null || ERROR "Path not found"
}

#######################################
# Pull a kubeconfig to instance and DO spaces and test via kubectl
# Globals:
#   CLUSTER_FULLNAME
# Arguments:
#   None
# Outputs:
#   Writes progress status
#######################################
function digitalocean::managed-kubernetes::pull_kubeconfig {
    DEBUG "Place a kubeconfig to executor instance and DO Spaces and test it via kubectl"
    local WAIT_TIMEOUT=5

    # Export to env variables
    export KUBECONFIG="$PRJ_ROOT/terraform/digitalocean/k8s/kubeconfig_$CLUSTER_FULLNAME"
    # Copy config to DO spaces (s3)
    run_cmd "s3cmd put '$PRJ_ROOT/terraform/digitalocean/k8s/kubeconfig_$CLUSTER_FULLNAME' \
            's3://${CLUSTER_FULLNAME}/kubeconfig_$CLUSTER_FULLNAME' \
            --host='$cluster_cloud_region.digitaloceanspaces.com' \
            --host-bucket='%(bucket)s.$cluster_cloud_region.digitaloceanspaces.com'" "" false

    # Copy config to default location
    run_cmd "cp '$PRJ_ROOT/terraform/digitalocean/k8s/kubeconfig_$CLUSTER_FULLNAME' '$HOME/.kube/config' 2>/dev/null" "" false

    INFO "DO Managed Kubernetes Cluster: Waiting for the Kubernetes Cluster to get ready"
    until kubectl version --request-timeout=5s >/dev/null 2>&1; do
        DEBUG "Waiting ${WAIT_TIMEOUT}s"
        sleep $WAIT_TIMEOUT
    done
    INFO "DO Managed Kubernetes Cluster: Ready to use!"
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
function digitalocean::managed-kubernetes::pull_kubeconfig_once {
    DEBUG "Test available kubeconfig via kubectl"

    INFO "Copy kubeconfig to cluster.dev executor"
    export KUBECONFIG="$PRJ_ROOT/terraform/digitalocean/k8s/kubeconfig_$CLUSTER_FULLNAME"

    # Copy config to default location
    run_cmd "cp '$PRJ_ROOT/terraform/digitalocean/k8s/kubeconfig_$CLUSTER_FULLNAME' '$HOME/.kube/config' 2>/dev/null" "" false
    kubectl version --request-timeout=5s >/dev/null 2>&1
    return $?
}
