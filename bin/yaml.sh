#!/usr/bin/env bash

# shellcheck source=logging.sh
source "$PRJ_ROOT"/bin/logging.sh

# Based on https://gist.github.com/pkuczynski/8665367

#######################################
# Read YAML file from Bash script and print out each KV in format:
# main_sublevel_sublevel2=value
# Globals:
#   None
# Arguments:
#   yaml_file - File name
#   prefix - Variable prefix. Default ""
#######################################
# shellcheck disable=SC1003
function yaml::parse {
    local yaml_file=$1
    local prefix=$2
    local s
    local w
    local fs

    s='[[:space:]]*'
    w='[a-zA-Z0-9_.-]*'
    fs="$(echo @|tr @ '\034')"

    (
        sed -e '/- [^\“]'"[^\']"'.*: /s|\([ ]*\)- \([[:space:]]*\)|\1-\'$'\n''  \1\2|g' |

        sed -ne '/^--/s|--||g; s|\"|\\\"|g; s/[[:space:]]*$//g;' \
            -e "/#.*[\"\']/!s| #.*||g; /^#/s|#.*||g;" \
            -e "s|^\($s\)\($w\)$s:$s\"\(.*\)\"$s\$|\1$fs\2$fs\3|p" \
            -e "s|^\($s\)\($w\)${s}[:-]$s\(.*\)$s\$|\1$fs\2$fs\3|p" |

        awk -F"$fs" '{
            indent = length($1)/2;
            if (length($2) == 0) { conj[indent]="+";} else {conj[indent]="";}
            vname[indent] = $2;
            for (i in vname) {if (i > indent) {delete vname[i]}}
                if (length($3) > 0) {
                    vn=""; for (i=0; i<indent; i++) {vn=(vn)(vname[i])("_")}
                    printf("%s%s%s%s=(\"%s\")\n", "'"$prefix"'",vn, $2, conj[indent-1],$3);
                }
            }' |

        sed -e 's/_=/+=/g' |

        awk 'BEGIN {
                FS="=";
                OFS="="
            }
            /(-|\.).*=/ {
                gsub("-|\\.", "_", $1)
            }
            { print }'
    ) < "$yaml_file" || ERROR "File '$yaml_file' not found"
}

#######################################
# Create variables from yaml file
# Globals:
#   None
# Arguments:
#   yaml_file - Cluster manifest file name
#   prefix - Variable prefix. Default ""
#######################################
function yaml::create_variables {
    local yaml_file="$1"
    local prefix="$2"
    eval "$(yaml::parse "$yaml_file" "$prefix")"
}

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
    arr=(
        cluster_cloud_domain
        cluster_cloud_provider
        cluster_cloud_region
        cluster_cloud_vpc
        cluster_name
        cluster_provisioner_instanceType
        cluster_provisioner_type
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
