#!/usr/bin/env bash

# Parse YAML configs in .cluster-dev/*
# shellcheck source=bin/yaml.sh
source "$PRJ_ROOT"/bin/yaml.sh # provides parse_yaml and create_variables
source "$PRJ_ROOT"/bin/logging.sh # PSR-3 compliant logging
source "$PRJ_ROOT"/bin/common.sh
source "$PRJ_ROOT"/bin/aws_common.sh
source "$PRJ_ROOT"/bin/aws_minikube.sh
source "$PRJ_ROOT"/bin/argocd.sh


# Mandatory variables passed to container by config
readonly CLUSTER_CONFIG_PATH=$1
readonly CLOUD_USER=$2
readonly CLOUD_PASS=$3

# Detect Git hosting and set GIT_REPO_NAME, GIT_REPO_ROOT variables
detect_git_provider

# =========================================================================== #
#                                    MAIN                                     #
# =========================================================================== #

DEBUG "Starting job in repo: $GIT_REPO_NAME with arguments  \
    CLUSTER_CONFIG_PATH: $CLUSTER_CONFIG_PATH, CLOUD_USER: $CLOUD_USER"

# Writes information about used software
output_software_info

# Iterate trough provided manifests and reconcile clusters
MANIFESTS=$(find "$CLUSTER_CONFIG_PATH" -type f) || ERROR "Manifest file/folder can't be found"
DEBUG "Manifests: $MANIFESTS"

for CLUSTER_MANIFEST_FILE in $MANIFESTS; do
    NOTICE "Now run: $CLUSTER_MANIFEST_FILE"
    DEBUG "Path where start new cycle: $PWD"

    # Clean variables from previus cluster yaml parsing
    unset cluster_apps

    yaml::parse "$CLUSTER_MANIFEST_FILE"
    yaml::create_variables "$CLUSTER_MANIFEST_FILE"
    yaml::check_that_required_variables_exist "$CLUSTER_MANIFEST_FILE"

    # Cloud selection. Declared via yaml::create_variables()
    # shellcheck disable=SC2154
    case $cluster_cloud_provider in
    aws)

        DEBUG "Cloud Provider: AWS. Initializing access variables"
        # Define AWS credentials
        export AWS_ACCESS_KEY_ID=$CLOUD_USER
        export AWS_SECRET_ACCESS_KEY=$CLOUD_PASS
        export AWS_DEFAULT_REGION=$cluster_cloud_region
        export CLUSTER_PREFIX=$GIT_REPO_NAME # CLUSTER_PREFIX equals git organization/username could be changed in other repo

        # Define cluster full name
        CLUSTER_FULLNAME=$cluster_name-$(echo "$CLUSTER_PREFIX" | awk -F "/" '{print$1}')
        # make sure it is not larger than 63 symbols and lowercase
        CLUSTER_FULLNAME=$(echo "$CLUSTER_FULLNAME" | cut -c 1-63 | awk '{print tolower($0)}')
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

            # Deploy ArgoCD via Terraform
            aws::init_argocd   "$cluster_name" "$cluster_cloud_region" "$cluster_cloud_domain"

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

    gcp)
        DEBUG "Cloud Provider: Google"
        ;;

    azure)
        DEBUG "Cloud Provider: Azure"
        ;;

    digitalocean)
        DEBUG "Cloud Provider: Azure"
        ;;

    esac

done
