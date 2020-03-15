#!/usr/bin/env bash

source ./logging.sh


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
