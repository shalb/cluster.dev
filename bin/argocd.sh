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
