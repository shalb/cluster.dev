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

    DEBUG "Echo env variables requied to deploy"
    DEBUG "DIGITALOCEAN_TOKEN: $DIGITALOCEAN_TOKEN SPACES_ACCESS_KEY_ID: $SPACES_ACCESS_KEY_ID SPACES_SECRET_ACCESS_KEY: $SPACES_SECRET_ACCESS_KEY AWS_ACCESS_KEY_ID: $AWS_ACCESS_KEY_ID AWS_SECRET_ACCESS_KEY: $AWS_SECRET_ACCESS_KEY"

    cd "$PRJ_ROOT"/terraform/digitalocean/backend/ || ERROR "Path not found"

    # Check if bucket already exist by trying to import it
    if (digitalocean::is_do_spaces_bucket_exists "$cluster_cloud_region"); then
        INFO "Terraform DO_SPACES_BACKEND_BUCKET: $DO_SPACES_BACKEND_BUCKET already exist"
    else
        NOTICE "Terraform DO_SPACES_BACKEND_BUCKET: $DO_SPACES_BACKEND_BUCKET not exist. It is going to be created"
        run_cmd "terraform apply -auto-approve \
                    -var='region=$cluster_cloud_region' \
                    -var='do_spaces_backend_bucket=$DO_SPACES_BACKEND_BUCKET' \
                    -var='do_token=$DIGITALOCEAN_TOKEN' \
                    -var='access_id=$SPACES_ACCESS_KEY_ID' \
                    -var='spaces_secret_key=$SPACES_SECRET_ACCESS_KEY'"
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
    terraform import -var="region=$cluster_cloud_region" -var="do_spaces_backend_bucket=$DO_SPACES_BACKEND_BUCKET" digitalocean_spaces_bucket.terraform_state "$cluster_cloud_region","$DO_SPACES_BACKEND_BUCKET" >/dev/null 2>&1
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
    local cluster_cloud_vpc=$1
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
# Arguments:
#   cluster_cloud_vpc
#   cluster_name
#   cluster_cloud_region
# Outputs:
#   Writes progress status
#######################################
function digitalocean::destroy_vpc {
    DEBUG "Create a VPC or use existing defined"
    local cluster_cloud_vpc=$1
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

#######################################
# Destroy all resources in cluster
# Globals:
#   DO_SPACES_BACKEND_BUCKET
#   CLUSTER_FULLNAME
# Arguments:
#   cluster_cloud_region
#   cluster_cloud_domain
#   cluster_cloud_provisioner_version
# Outputs:
#   Writes progress status
#######################################
function digitalocean::destroy {
        case $cluster_cloud_provisioner_type in
        managed-kubernetes)
            DEBUG "Destroy: Provisioner: DigitalOcean Kubernetes"
            if digitalocean::managed-kubernetes::pull_kubeconfig_once ; then
                digitalocean::destroy_addons "$CLUSTER_FULLNAME" "$cluster_cloud_region" "$cluster_cloud_domain"
            fi

            digitalocean::managed-kubernetes::destroy_cluster \
                "$CLUSTER_FULLNAME" \
                "$cluster_cloud_region" \
                "$cluster_cloud_provisioner_version"
            digitalocean::destroy_vpc "$cluster_cloud_vpc" "$CLUSTER_FULLNAME" "$cluster_cloud_region"
            digitalocean::destroy_domain "$cluster_cloud_region" "$CLUSTER_FULLNAME" "$cluster_cloud_domain"
            digitalocean::destroy_do_spaces_bucket "$cluster_cloud_region"
        ;;
        # end of digitalocean kubernetes
        esac
}

#######################################
# Writes commands for user for get access to cluster
# Globals:
#   CLUSTER_FULLNAME
# Outputs:
#   Writes commands to get cluster's kubeconfig and ssh key
#######################################
function digitalocean::output_access_keys {
    DEBUG "Writes commands for user for get access to cluster"

    # TODO: Add output as part of output status. Add commit-back hook with instructions to .cluster.dev/README.md

    KUBECONFIG_DOWNLOAD_MESSAGE="\
Download and apply your kubeconfig using commands: \n\
s3cmd get s3://${CLUSTER_FULLNAME}/kubeconfig_${CLUSTER_FULLNAME} ~/.kube/kubeconfig_${CLUSTER_FULLNAME} --host-bucket='%(bucket)s.$cluster_cloud_region.digitaloceanspaces.com' --host='$cluster_cloud_region.digitaloceanspaces.com' \n\
export KUBECONFIG=~/.kube/kubeconfig_${CLUSTER_FULLNAME} \n\
kubectl get ns \n
"
    NOTICE "$KUBECONFIG_DOWNLOAD_MESSAGE"

    # Add output to GitHub Action Step "steps.reconcile.outputs.(kubeconfig)"
    echo "::set-output name=kubeconfig::${KUBECONFIG_DOWNLOAD_MESSAGE}"
}


#######################################
# Deploy K8s Addons via Terraform
# Globals:
#   $DO_SPACES_BACKEND_BUCKET
#   CLUSTER_FULLNAME
# Arguments:
#   cluster_name
#   cluster_cloud_region
#   cluster_cloud_domain
#   config_path
# Outputs:
#   Writes progress status
#######################################
function digitalocean::init_addons {
    DEBUG "Deploy Kubernetes Addons via Terraform"
    local cluster_name=$1
    local cluster_cloud_region=$2
    local cluster_cloud_domain=$3
    local config_path=${4:-"~/.kube/config"}
    local do_token=$DIGITALOCEAN_TOKEN

    cd "$PRJ_ROOT"/terraform/digitalocean/addons/ || ERROR "Path not found"

    INFO "Kubernetes Addons: Init Terraform configuration"
    run_cmd "terraform init \
                -backend-config='bucket=$DO_SPACES_BACKEND_BUCKET' \
                -backend-config='key=states/terraform-addons.state' \
                -backend-config='endpoint=$cluster_cloud_region.digitaloceanspaces.com'"

    run_cmd "terraform plan \
                -var='region=$cluster_cloud_region' \
                -var='cluster_cloud_domain=$cluster_cloud_domain' \
                -var='cluster_name=$CLUSTER_FULLNAME' \
                -var='config_path=$config_path' \
                -var='do_token=$do_token' \
                -input=false \
                -out=tfplan-addons"

    INFO "Kubernetes Addons: Installing/Reconciling"
    run_cmd "terraform apply -auto-approve -compact-warnings -input=false tfplan-addons"

    local output
    output=$(terraform output)
    ARGOCD_ACCESS="ArgoCD Credentials:\n\
${output//$'\n'/'\n'}" # newline characters shielding

    NOTICE "$ARGOCD_ACCESS"

    echo "::set-output name=argocd::${ARGOCD_ACCESS}"

    cd - >/dev/null || ERROR "Path not found"
}

#######################################
# Destroy Kubernetes Addons via Terraform
# Globals:
#   CLUSTER_FULLNAME
#   DO_SPACES_BACKEND_BUCKET
# Arguments:
#   cluster_name
#   cluster_cloud_region
#   cluster_cloud_domain
# Outputs:
#   Writes progress status
#######################################
function digitalocean::destroy_addons {
    DEBUG "Delete Kubernetes Addons via Terraform"
    local cluster_name=$1
    local cluster_cloud_region=$2
    local cluster_cloud_domain=$3

    cd "$PRJ_ROOT"/terraform/digitalocean/addons/ || ERROR "Path not found"

    INFO "Kubernetes Addons: Init Terraform configuration"
    run_cmd "terraform init \
                -backend-config='bucket=$DO_SPACES_BACKEND_BUCKET' \
                -backend-config='key=states/terraform-addons.state' \
                -backend-config='endpoint=$cluster_cloud_region.digitaloceanspaces.com'"

    INFO "Kubernetes Addons: Destroying"
    run_cmd "terraform destroy -auto-approve -compact-warnings \
                -var='cluster_cloud_domain=$cluster_cloud_domain' \
                -var='cluster_name=$cluster_name' \
                -var='region=$cluster_cloud_region'" "" "false"

    cd - >/dev/null || ERROR "Path not found"
}
