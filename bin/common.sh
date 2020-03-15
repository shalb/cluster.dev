#!/usr/bin/env bash

source ./logging.sh

#######################################
# Deploy CertManager via kubectl
# Globals:
#   None
# Arguments:
#   None
# Outputs:
#   Writes progress status
#######################################
function kube::deploy_apps {
    INFO "Setup addons."
    DEBUG "Deploy Addons via kubectl"
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
function kube::destroy_apps {
    INFO "Remove addons."
    DEBUG "Delete Addons via kubectl"
    run_cmd "kubectl delete -f 'https://raw.githubusercontent.com/shalb/terraform-aws-minikube/8a147f7c0044c318ec37990b50f0cabb205e9b44/addons/letsencrypt-prod.yaml' || true"
    run_cmd "kubectl delete -f 'https://github.com/jetstack/cert-manager/releases/download/v0.13.0/cert-manager-no-webhook.yaml' || true"
    run_cmd "kubectl delete -f 'https://raw.githubusercontent.com/shalb/terraform-aws-minikube/master/addons/ingress.yaml' || true"
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
function output_software_info {
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
