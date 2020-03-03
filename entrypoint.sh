#!/bin/bash

# Parse YAML configs in .cluster-dev/*
source ./bin/yaml.sh # provides parse_yaml and create_variables

# Variables passed by Github Workflow to Action
CLUSTER_CONFIG_PATH=$1
CLOUD_USER=$2
CLOUD_PASS=$3
# For local testing run: ./entrypoint.sh .cluster.dev/minikube-one.yaml AWSUSER AWSPASS
#

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
export CLUSTER_PREFIX=$GITHUB_REPOSITORY # CLUSTER_PREFIX equals git organisation/username could be changed in other repo

# create uniqe s3 bucket from repo name and cluster name
S3_BACKEND_BUCKET=$(echo $CLUSTER_PREFIX|awk -F "/" '{print$1}')-$cluster_name
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
# TODO: implement switch for domain. https://github.com/shalb/cluster.dev/issues/2
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

#### Create a VPC or use existing defined ####
# KEY: vpc
# Possible options:
# default - use default vpc subnet
# create - create new vpc by terraform
# vpc-id - use client vpc, first subnet in a list

cluster_cloud_vpc_id=""
case ${cluster_cloud_vpc} in
    default|"")
        echo "*** Using default VPC"
        ;;
    create)
        # Create new VPC and get ID.
        echo "*** Creating new VPC"
        cd ../vpc/
        terraform init -backend-config="bucket=$S3_BACKEND_BUCKET" \
                  -backend-config="key=$cluster_name/terraform-vpc.state" \
                  -backend-config="region=$cluster_cloud_region"
        terraform plan \
                  -var="region=$cluster_cloud_region" \
                  -var="cluster_name=$CLUSTER_FULLNAME" \
                  -input=false \
                  -out=tfplan
        terraform apply -auto-approve -compact-warnings -input=false tfplan
        # Get VPC ID for later use.
        cluster_cloud_vpc_id=$(terraform output vpc_id)
        ;;
    *)
        # Use client VPC ID.
        echo "*** Using VPC ID ${cluster_cloud_vpc}"
        cluster_cloud_vpc_id=${cluster_cloud_vpc}
        ;;
esac

# Provisioner selection
case $cluster_provisioner_type in
minikube)
echo "*** Provisioner: Minikube"

## Deploy main Terraform code
echo "*** Init Terraform code with s3 backend"
cd ../minikube/
terraform init -backend-config="bucket=$S3_BACKEND_BUCKET" \
               -backend-config="key=$cluster_name/terraform.state" \
               -backend-config="region=$cluster_cloud_region" \

# TODO Minikube module is using Centos7 image which requires to be accepted and subscribed in MarketPlace:
# To do so please visit https://aws.amazon.com/marketplace/pp?sku=aw0evgkw8e5c1q413zgy5pjce

echo "*** Apply Terraform code execution"
terraform plan \
                  -var="region=$cluster_cloud_region" \
                  -var="cluster_name=$CLUSTER_FULLNAME" \
                  -var="aws_instance_type=$cluster_provisioner_instanceType" \
                  -var="hosted_zone=$cluster_cloud_domain" \
                  -var="vpc_id=$cluster_cloud_vpc_id" \
                  -input=false \
                  -out=tfplan

terraform apply -auto-approve -compact-warnings -input=false tfplan
## End of Deploy Minikube

# Pull a kubeconfig
function pull_kubeconfig {
  WAIT_TIMEOUT=5;
  until kubectl version --request-timeout=5s > /dev/null; do
      sleep $WAIT_TIMEOUT;
      aws s3 cp s3://${CLUSTER_FULLNAME}/kubeconfig_${CLUSTER_FULLNAME} ~/.kube/kubeconfig_${CLUSTER_FULLNAME}
      export KUBECONFIG=~/.kube/kubeconfig_${CLUSTER_FULLNAME}
      cp ~/.kube/kubeconfig_${CLUSTER_FULLNAME} ~/.kube/config > /dev/null
      echo "*** Waiting $WAIT_TIMEOUT seconds for Kubernetes Cluster gets ready";
  done
}

pull_kubeconfig

# Deploy CertManager
kubectl apply -f  "https://github.com/jetstack/cert-manager/releases/download/v0.13.0/cert-manager-no-webhook.yaml"
kubectl apply -f  "https://raw.githubusercontent.com/shalb/terraform-aws-minikube/8a147f7c0044c318ec37990b50f0cabb205e9b44/addons/letsencrypt-prod.yaml"

## Deploy ArgoCD
echo -e "${PURPLE}*** Installing/Reconciling ArgoCD...."
cd ../argocd/
terraform init -backend-config="bucket=$S3_BACKEND_BUCKET" \
               -backend-config="key=$cluster_name/terraform-argocd.state" \
               -backend-config="region=$cluster_cloud_region"

echo "*** Apply Terraform code execution....."
terraform plan \
               -var="argo_domain=argo-$CLUSTER_FULLNAME.$cluster_cloud_domain" \
               -input=false -out=tfplan-argocd

terraform apply -auto-approve -compact-warnings -input=false tfplan-argocd


## Apply output for user
# TODO Add output as part of output status. Add commit-back hook with instructions to .cluster.dev/README.md
PURPLE='\033[0;35m'
echo -e "${PURPLE}*** Download and apply your kubeconfig using commands:
aws s3 cp s3://${CLUSTER_FULLNAME}/kubeconfig_${CLUSTER_FULLNAME} ~/.kube/kubeconfig_${CLUSTER_FULLNAME}
export KUBECONFIG=\$KUBECONFIG:~/.kube/kubeconfig_${CLUSTER_FULLNAME}
kubectl get ns
"

echo -e "${PURPLE}*** Download your bastion ssh key using commands:
aws s3 cp s3://${CLUSTER_FULLNAME}/id_rsa_${CLUSTER_FULLNAME}.pem ~/.ssh/id_rsa_${CLUSTER_FULLNAME}.pem && chmod 600 ~/.ssh/id_rsa_${CLUSTER_FULLNAME}.pem
ssh -i ~/.ssh/id_rsa_${CLUSTER_FULLNAME}.pem centos@$CLUSTER_FULLNAME.$cluster_cloud_domain
"

## Output software versions
helmfile -v
kubectl version
git --version
aws --version


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

digitalocean)
echo "*** Cloud Provider Azure"
;;

esac

done

echo ::set-output name=status::\ exit_status=$?
