#!/bin/bash
# Parse YAML configs in .cluster-dev/*
source ./bin/yaml.sh # provides parse_yaml and create_variables

# Variables passed by Github Workflow to Action
CLUSTER_CONFIG_PATH=$1
CLOUD_USER=$2
CLOUD_PASS=$3
# For local testing run: ./entrypoint.sh .cluster.dev/minikube-one.yaml AWSUSER AWSPASS

echo "Starting job in repo: $GITHUB_REPOSITORY with arguments  \
      CLUSTER_CONFIG_PATH: $CLUSTER_CONFIG_PATH, CLOUD_USER: $CLOUD_USER" 

# Iterate trough provided manifests and reconcile clusters
for CLUSTER_MANIFEST_FILE in $(find $CLUSTER_CONFIG_PATH -type f); do

parse_yaml $CLUSTER_MANIFEST_FILE
create_variables $CLUSTER_MANIFEST_FILE

case $cluster_cloud_provider in
aws)

echo "Cloud Provider AWS. Initing access variables"
# Define AWS credentials
export AWS_ACCESS_KEY_ID=$CLOUD_USER
export AWS_SECRET_ACCESS_KEY=$CLOUD_PASS
export AWS_DEFAULT_REGION=$cluster_cloud_region
# create uniqe s3 bucket from repo name and cluster name
# TODO: ensure uniqnes: https://shalb.slack.com/archives/CRFSNTDGX/p1577644315023400
S3_BACKEND_BUCKET=$(echo $GITHUB_REPOSITORY|awk -F "/" '{print$1}')-$cluster_name
# make sure it is not larger than 63 symbols 
S3_BACKEND_BUCKET=$(echo $S3_BACKEND_BUCKET| cut -c 1-63)

# Create and init backend.
cd terraform/aws/backend/
terraform init
# Check if bucket already exist by trying to import it
if ( terraform import -var="region=$cluster_cloud_region" -var="s3_backend_bucket=$S3_BACKEND_BUCKET" aws_s3_bucket.terraform_state $S3_BACKEND_BUCKET ); then
echo "Terraform S3_BACKEND_BUCKET: $S3_BACKEND_BUCKET already exist";
else
echo "Terraform S3_BACKEND_BUCKET: $S3_BACKEND_BUCKET not exist. Creating one..."
terraform apply -auto-approve -var="region=$cluster_cloud_region" -var="s3_backend_bucket=$S3_BACKEND_BUCKET"
fi

case $cluster_provisioner_type in 
minikube)
echo "Provisioner: Minikube" 

# Deploy main Terraform code
echo "Init Terraform code with s3 backend"
cd ../minikube/
terraform init -backend-config="bucket=$S3_BACKEND_BUCKET" \
               -backend-config="key=$cluster_name/terraform.state" \
               -backend-config="region=$cluster_cloud_region" \

# HACK: need to add Publick key
mkdir ~/.ssh/
echo "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCi6UIVruH0CfKewYlSjA7oR6gjahZrkJ+k/0cj46nvYrORVcds2cijZPT34ACWkvXV8oYvXGWmvlGXV5H1sD0356zpjhRnGo6j4UZVS6KYX5HwObdZ6H/i+A9knEyXxOCyo6p4VeJIYGhVYcQT4GDAkxb8WXHVP0Ax/kUqrKx0a2tK9JjGkuLbufQc3yWhqcfZSVRU2a+M8f8EUmGLOc2VEi2mGoxVgikrelJ0uIGjLn63L6trrsbvasoBuILeXOAO1xICwtYFek/MexQ179NKqQ1Wx/+9Yx4Xc63MB0vR7kde6wxx2Auzp7CjJBFcSTz0TXSRsvF3mnUUoUrclNkr voa@auth.shalb.com" > ~/.ssh/id_rsa.pub

echo "Plan Terraform code execution"
terraform plan -compact-warnings -var="region=$cluster_cloud_region" -var="cluster_name=$cluster_name" -var="aws_instance_type=$cluster_provisioner_instanceType"


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