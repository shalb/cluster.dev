#!/bin/bash
# Parse YAML configs in .cluster-dev/*

echo "Hello $1"
time=$(date)
echo ::set-output name=time::$time

source yaml.sh

for CLUSTER_MANIFEST_FILE in $(find ../.cluster.dev/ -type f); do 
parse_yaml $CLUSTER_MANIFEST_FILE
create_variables $CLUSTER_MANIFEST_FILE

case $cluster_provisioner_type in
minikube)
echo "AWS Mikikube Password:" $cluster_clould_awsPassword
echo "terraform init && terraform plan" 
;;
eks)
echo "AWS EKS Password:" $cluster_clould_awsPassword
;;
esac
done