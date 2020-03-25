#!/usr/bin/env bash

# shellcheck source=logging.sh
source "$PRJ_ROOT"/bin/logging.sh

#######################################
# Deploy ArgoCD applications via kubectl
# Globals:
#   None
# Arguments:
#   cluster_apps
# Outputs:
#   Writes progress status
#######################################
function argocd::deploy_apps {
    local cluster_apps_array=( "${cluster_apps[@]}" )

    DEBUG "Deploy ArgoCD apps via kubectl from kubernetes/apps folder"
    DEBUG "Current folder $(pwd)"

    INFO "Deploy Apps from /kubernetes/apps/<folder> into ArgoCD"

    for ARGO_APP_DIR in "${cluster_apps_array[@]}"; do
        run_cmd "kubectl apply -f /kubernetes/apps/$ARGO_APP_DIR --recursive" "" "false";
        run_cmd "find /github/workspace"
    done

    #TODO: enable deletion from ArgoCD application that are installed but not mentioned in target folders manifests

}
