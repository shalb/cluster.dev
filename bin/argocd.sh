#!/usr/bin/env bash

source /bin/logging.sh


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

    #TODO: enable deletion from ArgoCD application that are installed but not mentioned in target folders manifests

}
