#!/usr/bin/env bash

# shellcheck source=logging.sh
source "$PRJ_ROOT"/bin/logging.sh
source "$PRJ_ROOT"/bin/common.sh

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

        run_cmd "aws s3 cp 's3://${CLUSTER_FULLNAME}/kubeconfig_$CLUSTER_FULLNAME' '$HOME/.kube/kubeconfig_$CLUSTER_FULLNAME' 2>/dev/null" "" false
        run_cmd "cp '$HOME/.kube/kubeconfig_$CLUSTER_FULLNAME' '$HOME/.kube/config' 2>/dev/null" "" false
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
function aws::minikube::pull_kubeconfig_once {
    DEBUG "Pull a kubeconfig to instance via kubectl"
    local WAIT_TIMEOUT=5

    INFO "Copy kubeconfig to instance with Minikube"
    export KUBECONFIG=~/.kube/kubeconfig_${CLUSTER_FULLNAME}
    run_cmd "aws s3 cp 's3://${CLUSTER_FULLNAME}/kubeconfig_$CLUSTER_FULLNAME' '$HOME/.kube/kubeconfig_$CLUSTER_FULLNAME' 2>/dev/null" "" "false"
    run_cmd "cp '$HOME/.kube/kubeconfig_$CLUSTER_FULLNAME' '$HOME/.kube/config' 2>/dev/null" "" "false"
    kubectl version --request-timeout=5s >/dev/null 2>&1
    return $?
}

#######################################
# Deploy Minikube cluster via Terraform
# Globals:
#   S3_BACKEND_BUCKET
# Arguments:
#   cluster_name
#   cluster_cloud_region
#   cluster_cloud_domain
#   cluster_cloud_provisioner_instanceType
# Outputs:
#   Writes progress status
#######################################
function aws::minikube::deploy_cluster {
    DEBUG "Deploy Minikube cluster via Terraform"
    local cluster_name=$1
    local region=$2
    local cluster_cloud_domain=$3
    local aws_instance_type=$4

    cd "$PRJ_ROOT"/terraform/aws/minikube/ || ERROR "Path not found"

    # Deploy main Terraform code
    INFO "Minikube cluster: Initializing Terraform configuration"
    run_cmd "terraform init \
                -backend-config='bucket=$S3_BACKEND_BUCKET' \
                -backend-config='key=states/terraform-k8s.state' \
                -backend-config='region=$cluster_cloud_region'"


    run_cmd "terraform plan \
                -var='cluster_name=$cluster_name' \
                -var='region=$region' \
                -var='hosted_zone=$cluster_name.$cluster_cloud_domain' \
                -var='aws_instance_type=$aws_instance_type' \
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
# Arguments:
#   cluster_name
#   cluster_cloud_region
#   cluster_cloud_domain
#   cluster_cloud_provisioner_instanceType
# Outputs:
#   Writes progress status
#######################################
function aws::minikube::destroy_cluster {
    DEBUG "Destroy Minikube cluster via Terraform"
    local cluster_name=$1
    local region=$2
    local cluster_cloud_domain=$3
    local aws_instance_type=$4

    cd "$PRJ_ROOT"/terraform/aws/minikube/ || ERROR "Path not found"

    # Deploy main Terraform code
    INFO "Minikube cluster: Initializing Terraform configuration"
    run_cmd "terraform init \
                -backend-config='bucket=$S3_BACKEND_BUCKET' \
                -backend-config='key=states/terraform-k8s.state' \
                -backend-config='region=$cluster_cloud_region'"


    INFO "Minikube cluster: Destroying "
    run_cmd "terraform destroy -auto-approve -compact-warnings \
                -var='cluster_name=$cluster_name' \
                -var='region=$region' \
                -var='hosted_zone=$cluster_name.$cluster_cloud_domain' \
                -var='aws_instance_type=$aws_instance_type'"

    cd - >/dev/null || ERROR "Path not found"
}
