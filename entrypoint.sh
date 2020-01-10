#!/bin/bash
# Parse YAML configs in .cluster-dev/*
source ./bin/yaml.sh # provides parse_yaml and create_variables

# Variables passed by Github Workflow to Action
CLUSTER_CONFIG_PATH=$1
CLOUD_USER=$2
CLOUD_PASS=$3
# For local testing run: ./entrypoint.sh .cluster.dev/minikube-one.yaml AWSUSER AWSPASS

echo "*** Starting job in repo: $GITHUB_REPOSITORY with arguments  \
      CLUSTER_CONFIG_PATH: $CLUSTER_CONFIG_PATH, CLOUD_USER: $CLOUD_USER" 

# Iterate trough provided manifests and reconcile clusters
for CLUSTER_MANIFEST_FILE in $(find $CLUSTER_CONFIG_PATH -type f); do

parse_yaml $CLUSTER_MANIFEST_FILE
create_variables $CLUSTER_MANIFEST_FILE

# Cloud selection
case $cluster_cloud_provider in
aws)

echo "*** Cloud Provider AWS. Initing access variables"
# Define AWS credentials
export AWS_ACCESS_KEY_ID=$CLOUD_USER
export AWS_SECRET_ACCESS_KEY=$CLOUD_PASS
export AWS_DEFAULT_REGION=$cluster_cloud_region

# create uniqe s3 bucket from repo name and cluster name
# TODO: ensure uniqnes: https://shalb.slack.com/archives/CRFSNTDGX/p1577644315023400
# TODO: Implement CLUSTER_PREFIX instead of GITHUB_REPOSITORY 
S3_BACKEND_BUCKET=$(echo $GITHUB_REPOSITORY|awk -F "/" '{print$1}')-$cluster_name
# make sure it is not larger than 63 symbols 
S3_BACKEND_BUCKET=$(echo $S3_BACKEND_BUCKET| cut -c 1-63)
# The same name would be used for domains
CLUSTER_FULLNAME=$S3_BACKEND_BUCKET

# Create and init backend.
cd terraform/aws/backend/
terraform init
# Check if bucket already exist by trying to import it
if ( terraform import -var="region=$cluster_cloud_region" -var="s3_backend_bucket=$S3_BACKEND_BUCKET" aws_s3_bucket.terraform_state $S3_BACKEND_BUCKET ); then
echo "*** Terraform S3_BACKEND_BUCKET: $S3_BACKEND_BUCKET already exist";
else
echo "*** Terraform S3_BACKEND_BUCKET: $S3_BACKEND_BUCKET not exist. Creating one..."
terraform apply -auto-approve -var="region=$cluster_cloud_region" -var="s3_backend_bucket=$S3_BACKEND_BUCKET"
fi

# Create a DNS domains/records if required
# TODO: implement switch for domain
if [ -z $cluster_cloud_domain ] ; then 
echo "*** The cluster domain is unset. Creating default one"
#cd ../route53/
#terraform init -backend-config="bucket=$S3_BACKEND_BUCKET" \
#               -backend-config="key=$cluster_name/terraform.state" \
#               -backend-config="region=$cluster_cloud_region" 
#terraform plan -compact-warnings \
#                  -var="region=$cluster_cloud_region" \
#                  -var="cluster_fullname=$CLUSTER_FULLNAME" \
#                  -var="cluster_domain=$cluster_cloud_domain"               
               
else
echo "*** The cluster domain is defined. So applying Terraform configuration for it"
#cd ../route53/
#terraform init -backend-config="bucket=$S3_BACKEND_BUCKET" \
#               -backend-config="key=$cluster_name/terraform.state" \
#               -backend-config="region=$cluster_cloud_region" 
#terraform plan -compact-warnings \
#                  -var="region=$cluster_cloud_region" \
#                  -var="cluster_fullname=$CLUSTER_FULLNAME" \
#                  -var="cluster_domain=$cluster_cloud_domain"
fi

# Create a VPC or use existing defined 
# TODO: implement switch for VPC
if [ -z $cluster_cloud_vpc ] ; then 
echo "*** The VPC is unset. Using default one"
else
echo "*** The VPC is defined. Applying Terraform configuration for VPC"
#cd ../vpc/
fi

# Provisioner selection
case $cluster_provisioner_type in 
minikube)
echo "*** Provisioner: Minikube" 

# Deploy main Terraform code
echo "*** Init Terraform code with s3 backend"
cd ../minikube/
terraform init -backend-config="bucket=$S3_BACKEND_BUCKET" \
               -backend-config="key=$cluster_name/terraform.state" \
               -backend-config="region=$cluster_cloud_region" \

# TODO HACK: need to add Publick keys. Move to terraform state
#mkdir ~/.ssh/
#echo "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCi6UIVruH0CfKewYlSjA7oR6gjahZrkJ+k/0cj46nvYrORVcds2cijZPT34ACWkvXV8oYvXGWmvlGXV5H1sD0356zpjhRnGo6j4UZVS6KYX5HwObdZ6H/i+A9knEyXxOCyo6p4VeJIYGhVYcQT4GDAkxb8WXHVP0Ax/kUqrKx0a2tK9JjGkuLbufQc3yWhqcfZSVRU2a+M8f8EUmGLOc2VEi2mGoxVgikrelJ0uIGjLn63L6trrsbvasoBuILeXOAO1xICwtYFek/MexQ179NKqQ1Wx/+9Yx4Xc63MB0vR7kde6wxx2Auzp7CjJBFcSTz0TXSRsvF3mnUUoUrclNkr voa@auth.shalb.com" > ~/.ssh/id_rsa.pub

# TODO Minikube module is using Centos7 image which requires to be accepted and subscribed in MarketPlace:
# To do so please visit https://aws.amazon.com/marketplace/pp?sku=aw0evgkw8e5c1q413zgy5pjce

echo "*** Apply Terraform code execution"
terraform apply -auto-approve -compact-warnings \
                  -var="region=$cluster_cloud_region" \
                  -var="cluster_name=$CLUSTER_FULLNAME" \
                  -var="aws_instance_type=$cluster_provisioner_instanceType" \
                  -var="hosted_zone=$cluster_cloud_domain"

# Apply output for user
PURPLE='\033[0;35m'
echo -e "${PURPLE}*** Download and apply your kubeconfig using commands: 
${PURPLE}aws s3 cp s3://${CLUSTER_FULLNAME}/kubeconfig_${CLUSTER_FULLNAME} ~/.kube/kubeconfig_${CLUSTER_FULLNAME} 
${PURPLE}export KUBECONFIG=\$KUBECONFIG:~/.kube/kubeconfig_${CLUSTER_FULLNAME}
${PURPLE}kubectl get ns

"

echo -e "${PURPLE}*** Download your bastion ssh key using commands: 
${PURPLE}aws s3 cp s3://${CLUSTER_FULLNAME}/id_rsa_${CLUSTER_FULLNAME}.pem ~/.ssh/id_rsa_${CLUSTER_FULLNAME}.pem && chmod 600 ~/.ssh/id_rsa_${CLUSTER_FULLNAME}.pem
${PURPLE}ssh -i ~/.ssh/id_rsa_${CLUSTER_FULLNAME}.pem centos@$CLUSTER_FULLNAME.$cluster_cloud_domain"

;; # end of minikube


eks)
echo "*** Cloud Provider AWS. Provisioner: EKS" 
;;
esac
;;

gcp)
echo "*** Cloud Provider Google"
;;

azure)
echo "*** Cloud Provider Azure"
;;

esac

done

echo ::set-output name=status::\ exit_status=$?  