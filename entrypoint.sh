#!/usr/bin/env bash

# Parse YAML configs in .cluster-dev/*
# shellcheck source=bin/yaml.sh
source "$PRJ_ROOT"/bin/yaml.sh # provides parse_yaml and create_variables
source "$PRJ_ROOT"/bin/logging.sh # PSR-3 compliant logging
source "$PRJ_ROOT"/bin/common.sh
source "$PRJ_ROOT"/bin/aws_common.sh
source "$PRJ_ROOT"/bin/aws_minikube.sh
source "$PRJ_ROOT"/bin/argocd.sh


# Variables passed by Github Workflow to Action
readonly CLUSTER_CONFIG_PATH=$1
readonly CLOUD_USER=$2
readonly CLOUD_PASS=$3
# For local testing run: ./entrypoint.sh .cluster.dev/minikube-one.yaml AWSUSER AWSPASS
#


# =========================================================================== #
#                                    MAIN                                     #
# =========================================================================== #


DEBUG "Starting job in repo: $GITHUB_REPOSITORY with arguments  \
    CLUSTER_CONFIG_PATH: $CLUSTER_CONFIG_PATH, CLOUD_USER: $CLOUD_USER"

# Writes information about used software
output_software_info

# Iterate trough provided manifests and reconcile clusters
MANIFESTS=$(find "$CLUSTER_CONFIG_PATH" -type f) || ERROR "Manifest file/folder can't be found"
DEBUG "Manifests: $MANIFESTS"

for CLUSTER_MANIFEST_FILE in $MANIFESTS; do
    DEBUG "Now run: $CLUSTER_MANIFEST_FILE"

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
        export CLUSTER_PREFIX=$GITHUB_REPOSITORY # CLUSTER_PREFIX equals git organization/username could be changed in other repo

        # create unique s3 bucket from repo name and cluster name
        S3_BACKEND_BUCKET=$(echo "$CLUSTER_PREFIX" | awk -F "/" '{print$1}')-$cluster_name
        # make sure it is not larger than 63 symbols and lowercase
        S3_BACKEND_BUCKET=$(echo "$S3_BACKEND_BUCKET" | cut -c 1-63 | awk '{print tolower($0)}')
        # The same name would be used for domains
        readonly CLUSTER_FULLNAME=$S3_BACKEND_BUCKET

        # Destroy if installed: false
        if [ "$cluster_installed" = "false" ]; then
            aws::destroy
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

            # Deploy k8s applications via kubectl
            kube::deploy_apps

            # Deploy ArgoCD via Terraform
            aws::init_argocd   "$cluster_name" "$cluster_cloud_region" "$cluster_cloud_domain"

            # Deploy ArgoCD apps via kubectl
            argocd::deploy_apps

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
