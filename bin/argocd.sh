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

    INFO "Deploy Apps from /kubernetes/apps/<folder> into ArgoCD"

    for i in "${cluster_apps[@]}"; do
    run_cmd "kubectl apply -f kubernetes/apps/$i --recursive";
    done

}
