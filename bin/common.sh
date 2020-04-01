#!/usr/bin/env bash

# shellcheck source=logging.sh
source "$PRJ_ROOT"/bin/logging.sh

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
    run_cmd "kubectl delete -f 'https://raw.githubusercontent.com/shalb/terraform-aws-minikube/master/addons/ingress.yaml' 2>/dev/null" "" false
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
