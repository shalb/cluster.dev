#!/usr/bin/env bash

# Parse YAML configs in .cluster-dev/*
# shellcheck source=bin/yaml.sh
source "$PRJ_ROOT"/bin/yaml.sh # provides parse_yaml and create_variables
source "$PRJ_ROOT"/bin/logging.sh # PSR-3 compliant logging
source "$PRJ_ROOT"/bin/common.sh
source "$PRJ_ROOT"/bin/aws_common.sh
source "$PRJ_ROOT"/bin/digitalocean_common.sh
source "$PRJ_ROOT"/bin/aws_minikube.sh
source "$PRJ_ROOT"/bin/argocd.sh


# Mandatory variables passed to container by config
readonly CLUSTER_CONFIG_PATH=${CLUSTER_CONFIG_PATH:-"./.cluster.dev/"}

# Detect Git hosting and set: GIT_PROVIDER, GIT_REPO_NAME, GIT_REPO_ROOT, CLUSTER_FULLNAME constants
detect_git_provider

# =========================================================================== #
#                                    MAIN                                     #
# =========================================================================== #

DEBUG "Starting job in repo: $GIT_REPO_NAME, CLUSTER_CONFIG_PATH: $CLUSTER_CONFIG_PATH"

# Writes information about used software
output_software_info

# Iterate trough provided manifests and reconcile clusters
MANIFESTS=$(find "$CLUSTER_CONFIG_PATH" -type f) || ERROR "Manifest file/folder can't be found"
DEBUG "Manifests: $MANIFESTS"

for CLUSTER_MANIFEST_FILE in $MANIFESTS; do
    NOTICE "Now run: $CLUSTER_MANIFEST_FILE"
    DEBUG "Path where start new cycle: $PWD"

    yaml::parse "$CLUSTER_MANIFEST_FILE"
    yaml::create_variables "$CLUSTER_MANIFEST_FILE"
    yaml::check_that_required_variables_exist "$CLUSTER_MANIFEST_FILE"

    # Cloud selection. Declared via yaml::create_variables()
    # shellcheck disable=SC2154
    case $cluster_cloud_provider in
    aws)

        DEBUG "Cloud Provider: AWS. Initializing access variables"
        # Define AWS credentials from ENV VARIABLES passed to container
        export AWS_ACCESS_KEY_ID=${AWS_ACCESS_KEY_ID}
        export AWS_SECRET_ACCESS_KEY=${AWS_SECRET_ACCESS_KEY}

        # Define full cluster name
        set_cluster_fullname
        # Define name for S3 bucket that would be user for terraform state
        S3_BACKEND_BUCKET=$CLUSTER_FULLNAME

        # Destroy if installed: false
        if [ "$cluster_installed" = "false" ]; then
            if (aws::is_s3_bucket_exists "$cluster_cloud_region"); then
                aws::destroy
            else
                DEBUG "S3 bucket ${S3_BACKEND_BUCKET} not exists. Nothing to destroy."
            fi
            continue
        fi

        # Create and init backend.
        # Check if bucket already exist by trying to import it
        aws::init_s3_bucket   "$cluster_cloud_region"

        # Create a DNS domains/records if required
        aws::init_route53   "$cluster_cloud_region" "$cluster_name" "$cluster_cloud_domain"

        # Create a VPC or use existing defined
        FUNC_RESULT=""
        aws::init_vpc   "$cluster_cloud_vpc" "$cluster_name" "$cluster_cloud_region"
        readonly CLUSTER_VPC_ID=${FUNC_RESULT}

        # Provisioner selection
        #
        case $cluster_provisioner_type in
        minikube)
            DEBUG "Provisioner: Minikube"

            # Deploy Minikube cluster via Terraform
            aws::minikube::deploy_cluster   "$cluster_name" "$cluster_cloud_region" "$cluster_provisioner_instanceType" "$cluster_cloud_domain" "$CLUSTER_VPC_ID"

            # Pull a kubeconfig to instance via kubectl
            aws::minikube::pull_kubeconfig

            # Deploy Kubernetes Addons via Terraform
            aws::init_addons   "$cluster_name" "$cluster_cloud_region" "$cluster_cloud_domain"

            # Deploy ArgoCD apps via kubectl
            argocd::deploy_apps   "$cluster_apps"

            # Writes commands for user for get access to cluster
            aws::output_access_keys   "$cluster_cloud_domain"
        ;;
        # end of minikube
        eks)
            DEBUG "Cloud Provider: AWS. Provisioner: EKS"
            ;;
        esac
        ;;

    digitalocean)

        DEBUG "Cloud Provider: DigitalOcean. Initializing access variables"
        # Define DO credentials from ENV VARIABLES passed to container
        export DIGITALOCEAN_TOKEN=${DIGITALOCEAN_TOKEN}
        export SPACES_ACCESS_KEY_ID=${SPACES_ACCESS_KEY_ID}
        export SPACES_SECRET_ACCESS_KEY=${SPACES_SECRET_ACCESS_KEY}

        # Define full cluster name
        set_cluster_fullname
        # Define name for S3 bucket that would be user for terraform state
        export DO_SPACES_BACKEND_BUCKET=$CLUSTER_FULLNAME

        # Create and init backend.
        # Check if bucket already exist by trying to import it
        digitalocean::init_do_spaces_bucket  "$cluster_cloud_region"

        case $cluster_provisioner_type in
        digitalocean-kubernetes)
            DEBUG "Provisioner: digitalocean-kubernetes"

        ;;
        esac


        ;;

    gcp)
        DEBUG "Cloud Provider: Google"
        ;;

    azure)
        DEBUG "Cloud Provider: Azure"
        ;;

    esac

done
