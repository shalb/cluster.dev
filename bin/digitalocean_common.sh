#!/usr/bin/env bash

# shellcheck source=logging.sh
source "$PRJ_ROOT"/bin/logging.sh

#######################################
# Create or use exiting DO Spaces bucket for Terraform states
# Globals:
#   DO_SPACES_BACKEND_BUCKET
# Arguments:
#   cluster_cloud_region
# Outputs:
#   Writes progress status
#######################################
function digitalocean::init_do_spaces_bucket {
    DEBUG "Create or use exiting DO spaces bucket for Terraform states"
    local cluster_cloud_region=$1

    cd "$PRJ_ROOT"/terraform/digitalocean/backend/ || ERROR "Path not found"

    # Check if bucket already exist by trying to import it
    if (digitalocean::is_do_spaces_bucket_exists "$cluster_cloud_region"); then
        INFO "Terraform DO_SPACES_BACKEND_BUCKET: $DO_SPACES_BACKEND_BUCKET already exist"
    else
        NOTICE "Terraform DO_SPACES_BACKEND_BUCKET: $DO_SPACES_BACKEND_BUCKET not exist. It is going to be created"
        run_cmd "terraform apply -auto-approve \
                    -var='region=$cluster_cloud_region' \
                    -var='do_spaces_backend_bucket=$DO_SPACES_BACKEND_BUCKET'"
    fi
    run_cmd "rm -rf *.tfstate"

    cd - >/dev/null || ERROR "Path not found"
}

#######################################
# Check if the do spaces bucket exists.
#
# Globals:
#   DO_SPACES_BACKEND_BUCKET
# Arguments:
#   cluster_cloud_region
# Outputs:
#   null
# Return:
#   exitcode
#######################################
function digitalocean::is_do_spaces_bucket_exists {
    local cluster_cloud_region=$1
    cd "$PRJ_ROOT"/terraform/digitalocean/backend/ || ERROR "Path not found"

    # Create and init backend.
    run_cmd "terraform init"

    terraform import -var="region=$cluster_cloud_region" -var="do_spaces_backend_bucket=$DO_SPACES_BACKEND_BUCKET" digitalocean_spaces_bucket.terraform_state `$cluster_cloud_region`,`$DO_SPACES_BACKEND_BUCKET` >/dev/null 2>&1
    return $?
}

#######################################
# Destroy DO Spaces bucket for Terraform states
# Globals:
#   DO_SPACES_BACKEND_BUCKET
# Arguments:
#   cluster_cloud_region
# Outputs:
#   Writes progress status
#######################################
function digitalocean::destroy_do_spaces_bucket {
    DEBUG "Destroy existing DO Spaces bucket for Terraform states. Bucket name: '${DO_SPACES_BACKEND_BUCKET}'"
    INFO "Destroying DO Spaces bucket for Terraform states."
    local cluster_cloud_region=$1
    run_cmd "s3cmd rb \"s3://${DO_SPACES_BACKEND_BUCKET}\" --force"
}

# Destroy all cluster.
function digitalocean::destroy {
        case $cluster_cloud_provisioner_type in
        digitalocean-kubernetes)
            DEBUG "Destroy: Provisioner: DigitalOcean Kubernetes"
            digitalocean::destroy_do_spaces_bucket "$cluster_cloud_region"
        ;;
        # end of digitalocean kubernetes
        esac
}
