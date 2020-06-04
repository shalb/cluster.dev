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

    INFO "Deploy Apps from ./kubernetes/apps/<folder> into ArgoCD"
    for (( i=0; i<${cluster_apps_count}; i++ ));
        do
            local cluster_app_folder="";
            eval cluster_app_folder='$'cluster_apps_$i;
            run_cmd "kubectl apply -f .$cluster_app_folder --recursive" "" "false";
            # Clean variables from previus cluster yaml parsing
        done
    #TODO: enable deletion from ArgoCD application that are installed but not mentioned in target folders manifests
}
