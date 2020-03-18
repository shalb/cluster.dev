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
function deploy_cert_manager {
    DEBUG "Deploy CertManager via kubectl"

    INFO "Setup TLS certificates"
    run_cmd "kubectl apply -f 'https://github.com/jetstack/cert-manager/releases/download/v0.13.0/cert-manager-no-webhook.yaml'"
    run_cmd "kubectl apply -f 'https://raw.githubusercontent.com/shalb/terraform-aws-minikube/8a147f7c0044c318ec37990b50f0cabb205e9b44/addons/letsencrypt-prod.yaml'"
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
