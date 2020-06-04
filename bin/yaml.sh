#!/usr/bin/env bash

# shellcheck source=logging.sh
source "$PRJ_ROOT"/bin/logging.sh

#######################################
# Check that cluster manifest have all needed KV
# Globals:
#   None
# Arguments:
#   file_path - full path to cluster manifest yaml file
#   required_vars - list of required variables
#######################################
function yaml::check_that_required_variables_exist {
    local file_path=$1
    # TODO: Add cloud provider support for required values
    arr=(
        cluster_cloud_domain
        cluster_cloud_provider
        cluster_cloud_region
        cluster_name
        cluster_cloud_provisioner_type
    )
    local required_vars=("${2:-${arr[@]}}")

    local have_err=false
    local one_indent="  "

    for var in "${required_vars[@]}"; do
        local yaml_structure=""
        local key
        # Replace underscores by yaml indent
        IFS='_' read -r -a array <<< "$var"
        for index in "${!array[@]}"; do
            # one_indent * index
            indent=$(printf "%${index}s");
            indent=${indent// /$one_indent}

            yaml_structure+="\n${indent}${array[index]}:"
            key=${array[index]}
        done

        # Get all vars
        var_exist="$( ( set -o posix ; set ) | grep "${var}=")"

        # Found and print all errors for file and only then stop execution
        if [ -z "$var_exist" ] ; then
            have_err=true

            DEBUG "Problem with '$var' KV. Get '$var_exist'"
            # TODO: Create documentation about all excepted types
            ERROR "Please, specify in '$file_path' required key '$key' here:
$yaml_structure <value>\n
For available values you can see our doc:
https://cluster.dev/" "%MESSAGE" "$LOG_DATE_FORMAT_SIMPLE" false
        fi
    done

    [ $have_err == true ] && exit 1
}
