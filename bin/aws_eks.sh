#!/usr/bin/env bash

# shellcheck source=logging.sh
source "$PRJ_ROOT"/bin/logging.sh

#######################################
# Deploy  cluster via Terraform
# Globals:
#   S3_BACKEND_BUCKET
#   CLUSTER_FULLNAME
# Arguments:
#   cluster_name
#   cluster_cloud_region
#   cluster_cloud_availability_zones
#   cluster_cloud_domain
#   cluster_cloud_vpc_id
#   cluster_cloud_provisioner_version
#   cluster_cloud_provisioner_node_group__name[@]
# Outputs:
#   Writes progress status
#######################################
function aws::eks::deploy_cluster {
    DEBUG "Deploy eks cluster via Terraform"
    local cluster_name=$1
    local region=$2
    local availability_zones=${3:-$cluster_cloud_region"a"} # if azs are not set we use 'a'-zone by default
    availability_zones=$(to_tf_list "$availability_zones") # convert to terraform list format
    local cluster_cloud_domain=$4
    local vpc_id=$5
    local cluster_version=${6:-"1.16"} # set default Kubernetes version
    local worker_additional_security_group_ids=${7:-[]}
    worker_additional_security_group_ids=$(to_tf_list "$worker_additional_security_group_ids")

# Generate worker_groups.tfvars with worker_groups_launch_template to pass module and create workers
            node_group_lenght=${#cluster_cloud_provisioner_node_group__name[@]}
                    echo "worker_groups_launch_template = [" > worker_groups.tfvars
            for (( i=0; i<${node_group_lenght}; i++ ));
                do
                    echo "{" >>  worker_groups.tfvars
                    echo    name                          = \""${cluster_cloud_provisioner_node_group__name[i]}"\" >> worker_groups.tfvars
                    echo    instance_type                 = \""${cluster_cloud_provisioner_node_group__instance_type[i]}"\" >> worker_groups.tfvars
                    echo    asg_desired_capacity          = "${cluster_cloud_provisioner_node_group__asg_desired_capacity[i]}" >> worker_groups.tfvars
                    echo    asg_max_size                  = "${cluster_cloud_provisioner_node_group__asg_max_size[i]}" >> worker_groups.tfvars
                    echo    asg_min_size                  = "${cluster_cloud_provisioner_node_group__asg_min_size[i]}" >> worker_groups.tfvars
                    echo    root_volume_size              = "${cluster_cloud_provisioner_node_group__root_volume_size[i]}" >> worker_groups.tfvars
                    echo    kubelet_extra_args            = \""${cluster_cloud_provisioner_node_group__kubelet_extra_args[i]}"\" >> worker_groups.tfvars
                    echo    override_instance_types       =  $(to_tf_list "${cluster_cloud_provisioner_node_group__instance_type_override[i]:-[]}") >> worker_groups.tfvars
                    echo    spot_allocation_strategy      = \""${cluster_cloud_provisioner_node_group__spot_allocation_strategy[i]}"\" >> worker_groups.tfvars
                    echo    spot_instance_pools           = "${cluster_cloud_provisioner_node_group__spot_instance_pools[i]:-10}" >> worker_groups.tfvars
                    echo    spot_max_price                = \""${cluster_cloud_provisioner_node_group__spot_max_price[i]}"\" >> worker_groups.tfvars
                    echo    on_demand_base_capacity       = "${cluster_cloud_provisioner_node_group__on_demand_base_capacity[i]:-0}" >> worker_groups.tfvars
                    echo    on_demand_percentage_above_base_capacity = "${cluster_cloud_provisioner_node_group__on_demand_percentage_above_base_capacity[i]:-0}" >> worker_groups.tfvars
                    echo "}," >>  worker_groups.tfvars
                done
                    echo "]" >> worker_groups.tfvars

    INFO "EKS Cluster: worker_groups.tfvars prepared"
    cat worker_groups.tfvars
    cp worker_groups.tfvars "$PRJ_ROOT"/terraform/aws/eks/

    cd "$PRJ_ROOT"/terraform/aws/eks/ || ERROR "Path not found"

    # Deploy main Terraform code
    INFO "EKS Cluster: Initializing Terraform configuration"
    run_cmd "terraform init \
                -backend-config='bucket=$S3_BACKEND_BUCKET' \
                -backend-config='key=$cluster_name/terraform.state' \
                -backend-config='region=$region'"


    run_cmd "terraform plan \
                -var='cluster_name=$CLUSTER_FULLNAME' \
                -var='region=$region' \
                -var='availability_zones=$availability_zones' \
                -var='vpc_id=$vpc_id' \
                -var='cluster_version=$cluster_version' \
                -var='worker_additional_security_group_ids=$worker_additional_security_group_ids' \
                -var-file='worker_groups.tfvars' \
                -input=false \
                -out=tfplan"

    INFO "EKS Cluster: Creating infrastructure"
    run_cmd "terraform apply -auto-approve -compact-warnings -input=false tfplan"

    cd - >/dev/null || ERROR "Path not found"
}

#######################################
# Destroy EKS cluster via Terraform
# Globals:
#   S3_BACKEND_BUCKET
#   CLUSTER_FULLNAME
# Arguments:
#   cluster_name
#   cluster_cloud_region
#   cluster_cloud_provisioner_instanceType
#   cluster_cloud_domain
# Outputs:
#   Writes progress status
#######################################
function aws::eks::destroy_cluster {
    DEBUG "Destroy eks cluster via Terraform"
    local cluster_name=$1
    local cluster_cloud_region=$2
    local cluster_cloud_provisioner_instanceType=$3
    local cluster_cloud_domain=$4

    cd "$PRJ_ROOT"/terraform/aws/eks/ || ERROR "Path not found"

    # Deploy main Terraform code
    INFO "eks cluster: Initializing Terraform configuration"
    run_cmd "terraform init \
                -backend-config='bucket=$S3_BACKEND_BUCKET' \
                -backend-config='key=$cluster_name/terraform.state' \
                -backend-config='region=$cluster_cloud_region'"


    INFO "eks cluster: Destroying "
    run_cmd "terraform destroy -auto-approve -compact-warnings \
                -var='region=$cluster_cloud_region' \
                -var='cluster_name=$CLUSTER_FULLNAME' \
                -var='aws_instance_type=$cluster_cloud_provisioner_instanceType' \
                -var='hosted_zone=$CLUSTER_FULLNAME.$cluster_cloud_domain'"

    cd - >/dev/null || ERROR "Path not found"
}
