#!/usr/bin/env bash

# shellcheck source=logging.sh
source "$PRJ_ROOT"/bin/logging.sh

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
