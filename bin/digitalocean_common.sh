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
    INFO "Importing DO Spaces bucket for Terraform states."
    terraform import -var="region=$cluster_cloud_region" -var="do_spaces_backend_bucket=$DO_SPACES_BACKEND_BUCKET" digitalocean_spaces_bucket.terraform_state "$cluster_cloud_region","$DO_SPACES_BACKEND_BUCKET"
    return $?
}

#######################################
# Create a DNS domains/records if required
# Globals:
#   DO_SPACES_BACKEND_BUCKET
# Arguments:
#   cluster_cloud_region
#   cluster_name
#   cluster_cloud_domain
#######################################
function digitalocean::init_domain {
    DEBUG "Create a DNS domains/records if required"
    local default_domain="cluster.dev"
    local cluster_cloud_region=$1
    local cluster_name=$2
    local cluster_cloud_domain=${3:-$default_domain}

    cd "$PRJ_ROOT"/terraform/digitalocean/domain/ || ERROR "Path not found"

    # Init terraform state for DNS
    INFO "DNS: Initializing Terraform configuration"
    run_cmd "terraform init \
                -backend-config='bucket=$DO_SPACES_BACKEND_BUCKET' \
                -backend-config='key=states/terraform-dns.state' \
                -backend-config='endpoint=$cluster_cloud_region.digitaloceanspaces.com'"

    # Create or update zone
    if [ "$cluster_cloud_domain" = "$default_domain" ]; then
        INFO "The cluster domain is unset. DNS sub-zone would be created in $default_domain"
        zone_delegation=true
    else
        INFO "The cluster domain defined. DNS sub-zone would be created in $cluster_cloud_domain"
        zone_delegation=false
    fi

    # Execute terraform
    run_cmd "terraform plan -compact-warnings \
            -var='region=$cluster_cloud_region' \
            -var='cluster_name=$cluster_name' \
            -var='cluster_domain=$cluster_cloud_domain' \
            -var='zone_delegation=$zone_delegation' \
            -input=false \
            -out=tfplan"

    run_cmd "terraform apply -auto-approve -compact-warnings -input=false tfplan"
    INFO "DNS Zone: $cluster_name.$cluster_cloud_domain has been created."

    cd - >/dev/null || ERROR "Path not found"
}

#######################################
# Destroy a DNS domains/records if required
# Globals:
#   DO_SPACES_BACKEND_BUCKET
# Arguments:
#   cluster_cloud_region
#   cluster_name
#   cluster_cloud_domain
#######################################
function digitalocean::destroy_domain {
    local default_domain="cluster.dev"
    local cluster_cloud_region=$1
    local cluster_name=$2
    local cluster_cloud_domain=${3:-$default_domain}

    # Init terraform state for DNS
    cd "$PRJ_ROOT"/terraform/digitalocean/domain/ || ERROR "Path not found"

    INFO "DNS: Initializing Terraform configuration"
    run_cmd "terraform init \
                -backend-config='bucket=$DO_SPACES_BACKEND_BUCKET' \
                -backend-config='key=states/terraform-dns.state' \
                -backend-config='endpoint=$cluster_cloud_region.digitaloceanspaces.com'"

    # Create or update zone
    if [ "$cluster_cloud_domain" = "$default_domain" ]; then
        zone_delegation=true
    else
        zone_delegation=false
    fi

    # Execute terraform
    INFO "Destroying a DNS zone $cluster_name.$cluster_cloud_domain"
    run_cmd "terraform  destroy -auto-approve  \
            -var='region=$cluster_cloud_region' \
            -var='cluster_domain=$cluster_cloud_domain' \
            -var='zone_delegation=$zone_delegation' \
            -var='cluster_name=$cluster_name'"

    INFO "DNS Zone: $cluster_name.$cluster_cloud_domain has been deleted."

    cd - >/dev/null || ERROR "Path not found"
}

#######################################
# Create a VPC or use existing defined
# Globals:
#   DO_SPACES_BACKEND_BUCKET
#   CLUSTER_FULLNAME
#   FUNC_RESULT
# Arguments:
#   cluster_cloud_vpc
#   cluster_name
#   cluster_cloud_region
#   cluster_cloud_vpc_cidr
# Outputs:
#   Writes progress status
# KEY: vpc (cluster_cloud_vpc)
# Possible options:
#   default - use default vpc subnet
#   create - create new vpc by terraform
#   vpc-id - use client vpc, first subnet in a list
#######################################
function digitalocean::init_vpc {
    DEBUG "Create a VPC or use existing defined"
    local cluster_cloud_vpc=${1:-"create"}
    local cluster_name=$2
    local cluster_cloud_region=$3
    local vpc_cidr=${4:-"10.8.0.0/18"} # set default VPC cidr

    cd "$PRJ_ROOT"/terraform/digitalocean/vpc/ || ERROR "Path not found"

    # Create/Init VPC and get ID.
    INFO "VPC: Initializing Terraform configuration"
    run_cmd "terraform init \
                -backend-config='bucket=$DO_SPACES_BACKEND_BUCKET' \
                -backend-config='key=states/terraform-vpc.state' \
                -backend-config='endpoint=$cluster_cloud_region.digitaloceanspaces.com'"

    run_cmd "terraform plan \
                -var='vpc_id=$cluster_cloud_vpc' \
                -var='cluster_name=$cluster_name' \
                -var='region=$cluster_cloud_region' \
                -var='ip_range=$vpc_cidr' \
                -input=false \
                -out=tfplan"

    INFO "VPC: Apply infrastructure changes"
    run_cmd "terraform apply -auto-approve -compact-warnings -input=false tfplan"

    cd - >/dev/null || ERROR "Path not found"
}

#######################################
# Destroy VPC
# Globals:
#   DO_SPACES_BACKEND_BUCKET
#   CLUSTER_FULLNAME
# Arguments:
#   cluster_cloud_vpc
#   cluster_name
#   cluster_cloud_region
# Outputs:
#   Writes progress status
#######################################
function digitalocean::destroy_vpc {
    DEBUG "Create a VPC or use existing defined"
    local cluster_cloud_vpc=${1:-"create"}
    local cluster_name=$2
    local cluster_cloud_region=$3

    DEBUG "Destroy created VPC keep default unchanged"
    cd "$PRJ_ROOT"/terraform/digitalocean/vpc/ || ERROR "Path not found"

            # Create/Init VPC and get ID.
            INFO "VPC: Initializing Terraform configuration"
            run_cmd "terraform init \
                        -backend-config='bucket=$DO_SPACES_BACKEND_BUCKET' \
                        -backend-config='key=states/terraform-vpc.state' \
                        -backend-config='endpoint=$cluster_cloud_region.digitaloceanspaces.com'"

            INFO "VPC: Destroying"
            run_cmd "terraform destroy -auto-approve -compact-warnings \
                        -var='vpc_id=$cluster_cloud_vpc' \
                        -var='cluster_name=$cluster_name' \
                        -var='region=$cluster_cloud_region'"

    cd - >/dev/null || ERROR "Path not found"
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
    run_cmd "s3cmd rb \"s3://${DO_SPACES_BACKEND_BUCKET}\" \
        --host='$cluster_cloud_region.digitaloceanspaces.com' \
        --host-bucket='%(bucket)s.$cluster_cloud_region.digitaloceanspaces.com' \
        --recursive \
        --force"
}

# Destroy all cluster.
function digitalocean::destroy {
        case $cluster_cloud_provisioner_type in
        managed-kubernetes)
            DEBUG "Destroy: Provisioner: DigitalOcean Kubernetes"
            digitalocean::managed-kubernetes::destroy_cluster \
                "$cluster_name" \
                "$cluster_cloud_region" \
                "$cluster_cloud_provisioner_version" \
                "$cluster_cloud_provisioner_nodeSize" \
                "$cluster_cloud_provisioner_minNodes" \
                "$cluster_cloud_provisioner_maxNodes"
            digitalocean::destroy_vpc "$cluster_cloud_vpc" "$cluster_name" "$cluster_cloud_region"
            digitalocean::destroy_domain "$cluster_cloud_region" "$CLUSTER_FULLNAME" "$cluster_cloud_domain"
            digitalocean::destroy_do_spaces_bucket "$cluster_cloud_region"
        ;;
        # end of digitalocean kubernetes
        esac
}
