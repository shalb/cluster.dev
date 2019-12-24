#!/bin/bash
# Parse YAML configs in .cluster-dev/*

CLUSTER_CONFIG_PATH=$1
CLOUD_USER=$2
CLOUD_PASS=$3

echo "Starting job with arguments CLUSTER_CONFIG_PATH: $CLUSTER_CONFIG_PATH, CLOUD_USER: $CLOUD_USER" 

source ./bin/yaml.sh

for CLUSTER_MANIFEST_FILE in $(find $CLUSTER_CONFIG_PATH -type f); do 
parse_yaml $CLUSTER_MANIFEST_FILE
create_variables $CLUSTER_MANIFEST_FILE

case $cluster_clould_provider in
aws)
echo "Cloud Provider AWS" 
terraform init 
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