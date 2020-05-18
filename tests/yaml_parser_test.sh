#!/usr/bin/env bash

source ../bin/yaml.sh # provides parse_yaml and create_variables
source ../bin/logging.sh # PSR-3 compliant logging

CLUSTER_CONFIG_PATH='../.cluster.dev/aws-eks.yaml'

for CLUSTER_MANIFEST_FILE in $(find "$CLUSTER_CONFIG_PATH" -type f  || ERROR "Manifest file/folder can't be found" && exit 1 ); do

    yaml::parse "$CLUSTER_MANIFEST_FILE"
    yaml::create_variables "$CLUSTER_MANIFEST_FILE"
    yaml::check_that_required_variables_exist "$CLUSTER_CONFIG_PATH/$CLUSTER_MANIFEST_FILE"


echo "---------------"
node_group_lenght=${#cluster_cloud_provisioner_node_group__name[@]}

echo "worker_groups_launch_template = ["
for (( i=0; i<${node_group_lenght}; i++ ));
    do
    echo "  {"
    echo  "    name                          = ${cluster_cloud_provisioner_node_group__name[i]}"
    echo  "    instance_type                 = ${cluster_cloud_provisioner_node_group__instance_type[i]}"
    echo  "    asg_desired_capacity          = ${cluster_cloud_provisioner_node_group__asg_desired_capacity[i]}"
    echo  "    asg_max_size                  = ${cluster_cloud_provisioner_node_group__asg_max_size[i]}"
    echo  "    asg_min_size                  = ${cluster_cloud_provisioner_node_group__asg_min_size[i]}"
    echo  "    root_volume_size              = ${cluster_cloud_provisioner_node_group__root_volume_size[i]}"
    echo  "    kubelet_extra_args            = ${cluster_cloud_provisioner_node_group__kubelet_extra_args[i]}"
    echo  "    additional_security_group_ids = ${cluster_cloud_provisioner_node_group__additional_security_group_ids[i]}"
    echo  "    override_instance_types       = ${cluster_cloud_provisioner_node_group__instance_type_override[i]}"
    echo  "    spot_allocation_strategy      = ${cluster_cloud_provisioner_node_group__spot_allocation_strategy[i]}"
    echo  "    spot_instance_pools           = ${cluster_cloud_provisioner_node_group__spot_instance_pools[i]}"
    echo  "    spot_max_price                = ${cluster_cloud_provisioner_node_group__spot_max_price[i]}"
    echo  "    on_demand_base_capacity       = ${cluster_cloud_provisioner_node_group__on_demand_base_capacity[i]}"
    echo  "    on_demand_percentage_above_base_capacity = ${cluster_cloud_provisioner_node_group__on_demand_percentage_above_base_capacity[i]}"
    echo  "    # base values"
    echo  "    subnets                       = var.subnets"
    echo  "    key_name                      = ${cluster_name}-bastion"
    echo  "    additional_userdata           = \"\""
    echo  "  },"

    done
    echo "]"
echo "-----------------"

echo "cluster apps to be installed:" "${cluster_apps[@]}"
    for i in "${cluster_apps[@]}"; do echo "$i"; done
done
