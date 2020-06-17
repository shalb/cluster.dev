#!/usr/bin/env bash

# Parse YAML configs in .cluster-dev/*
# shellcheck source=bin/yaml.sh
source "$PRJ_ROOT"/bin/yaml.sh # provides parse_yaml and create_variables
source "$PRJ_ROOT"/bin/logging.sh # PSR-3 compliant logging
source "$PRJ_ROOT"/bin/common.sh
source "$PRJ_ROOT"/bin/aws_common.sh
source "$PRJ_ROOT"/bin/digitalocean_common.sh
source "$PRJ_ROOT"/bin/aws_minikube.sh
source "$PRJ_ROOT"/bin/aws_eks.sh
source "$PRJ_ROOT"/bin/argocd.sh


# Mandatory variables passed to container by config
readonly CLUSTER_CONFIG_PATH=${CLUSTER_CONFIG_PATH:-"./.cluster.dev/"}

# Detect Git hosting and set: GIT_PROVIDER, GIT_REPO_NAME, CLUSTER_FULLNAME constants
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

    # Unset all env variables by `cluster_` that could be kept from previous run
    while read var; do unset $var; done < <(env | grep cluster_ | awk -F "=" '{print$1}')
    # Parse Yaml Manifest and export variables with https://github.com/shalb/yamltoenv
    source <(/bin/yamltoenv -f "$CLUSTER_MANIFEST_FILE")
    # Check required variables are set
    yaml::check_that_required_variables_exist "$CLUSTER_MANIFEST_FILE"

    # Define full cluster name
    FUNC_RESULT="";
    set_cluster_fullname "$cluster_name" "$GIT_REPO_NAME"
    CLUSTER_FULLNAME=${FUNC_RESULT}

    # Cloud selection. Declared via yaml::create_variables()
    # shellcheck disable=SC2154
    case $cluster_cloud_provider in
    aws)

        DEBUG "Cloud Provider: AWS. Initializing access variables"

        # Define name for S3 bucket that would be user for terraform state.
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

        # Create and init backend. Check if bucket already exist by trying to import it.
        aws::init_s3_bucket   "$cluster_cloud_region"

        # Create a DNS zone if required.
        aws::init_route53   "$cluster_cloud_region" "$CLUSTER_FULLNAME" "$cluster_cloud_domain"

        # Create a VPC or use existing defined.
        aws::init_vpc   "$cluster_cloud_vpc" "$CLUSTER_FULLNAME" "$cluster_cloud_region" "$cluster_cloud_availability_zones" "$cluster_cloud_vpc_cidr"

        # Provisioner selection.
        #
        case $cluster_cloud_provisioner_type in
        minikube)
            DEBUG "Provisioner: Minikube"
            # Deploy Minikube cluster via Terraform
            aws::minikube::deploy_cluster   "$CLUSTER_FULLNAME" "$cluster_cloud_region" "$cluster_cloud_domain" "$cluster_cloud_provisioner_instanceType"

            # Pull a kubeconfig to instance via kubectl
            aws::minikube::pull_kubeconfig

            # Deploy Kubernetes Addons via Terraform
            aws::init_addons   "$CLUSTER_FULLNAME" "$cluster_cloud_region" "$cluster_cloud_domain"
        ;;
        # end of minikube
        eks)
            DEBUG "Cloud Provider: AWS. Provisioner: EKS"
            # Deploy Minikube cluster via Terraform
            aws::eks::deploy_cluster    "$CLUSTER_FULLNAME" "$cluster_cloud_region" "$cluster_cloud_availability_zones" "$cluster_cloud_vpc" "$cluster_version"

            # Pull a kubeconfig to instance via kubectl
            aws::eks::pull_kubeconfig

            # Deploy Kubernetes Addons via Terraform
            aws::init_addons   "$CLUSTER_FULLNAME" "$cluster_cloud_region" "$cluster_cloud_domain" "$PRJ_ROOT/terraform/aws/eks/kubeconfig_$CLUSTER_FULLNAME" "true"

            ;;
        esac

            # Deploy ArgoCD apps via kubectl
            argocd::deploy_apps

            # Writes commands for user for get access to cluster
            aws::output_access_keys   "$cluster_cloud_domain"
        ;;

    digitalocean)

        DEBUG "Cloud Provider: DigitalOcean. Initializing access variables"
        # Define DO credentials from ENV VARIABLES passed to container
        # TODO: Check that DIGITALOCEAN_TOKEN SPACES_ACCESS_KEY_ID SPACES_SECRET_ACCESS_KEY are set

        # s3cmd DO remove bucket ENV VARIABLES
        AWS_ACCESS_KEY_ID=${SPACES_ACCESS_KEY_ID}
        AWS_SECRET_ACCESS_KEY=${SPACES_SECRET_ACCESS_KEY}

        # Define full cluster name
        FUNC_RESULT="";
        set_cluster_fullname "$cluster_name" "$GIT_REPO_NAME"
        CLUSTER_FULLNAME=${FUNC_RESULT}

        # Destroy if installed: false
        if [ "$cluster_installed" = "false" ]; then
            if (digitalocean::is_do_spaces_bucket_exists "$cluster_cloud_region"); then
                digitalocean::destroy
            else
                DEBUG "S3 Spaces bucket ${S3_BACKEND_BUCKET} not exists. Nothing to destroy."
            fi
            continue
        fi

        # Define name for S3 bucket that would be user for terraform state
        DO_SPACES_BACKEND_BUCKET=$CLUSTER_FULLNAME

        # Create and init backend.
        # Check if bucket already exist by trying to import it
        digitalocean::init_do_spaces_bucket  "$cluster_cloud_region"

        case $cluster_cloud_provisioner_type in
        managed-kubernetes)
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
