#!/bin/bash
# Parse YAML configs in .cluster-dev/*
source ./bin/yaml.sh

CLUSTER_CONFIG_PATH=$1
CLOUD_USER=$2
CLOUD_PASS=$3

echo "Starting job in repo: $GITHUB_REPOSITORY with arguments CLUSTER_CONFIG_PATH: $CLUSTER_CONFIG_PATH, CLOUD_USER: $CLOUD_USER" 

for CLUSTER_MANIFEST_FILE in $(find $CLUSTER_CONFIG_PATH -type f); do

parse_yaml $CLUSTER_MANIFEST_FILE
create_variables $CLUSTER_MANIFEST_FILE

case $cluster_cloud_provider in
aws)

case $cluster_cloud_provisioner in 
minikube)
cd terraform/aws/minikube/ && ls -las
echo "Cloud Provider AWS. Provisioner: Minikube" 
echo terraform init \
echo -backend-config="bucket=$GITHUB_REPOSITORY" \
echo -backend-config="$cluster_name/terraform.state" \
echo -backend-config="region=$cluster_cloud_region"
pwd
terraform init  \
-backend-config="bucket=$GITHUB_REPOSITORY" \
-backend-config="$cluster_name/terraform.state" \
-backend-config="region=$cluster_cloud_region"
-backend-config="access_key=$CLOUD_USER"
-backend-config="secret_key=$CLOUD_PASS"
;;

eks)
echo "Cloud Provider AWS. Provisioner: EKS" 
;;
esac
;;

gcp)
echo "Cloud Provider Google"
;;

azure)
echo "Cloud Provider Azure"
;;

esac

done

echo ::set-output name=status::\ exit_status=$?  