#!/usr/bin/env bash

source ../bin/common.sh

    CLUSTER_MANIFEST_FILE='../.cluster.dev/aws-eks.yaml'
    ~/yamltoenv -f "$CLUSTER_MANIFEST_FILE"
    source <(~/yamltoenv -f "$CLUSTER_MANIFEST_FILE")

if [ $cluster_cloud_vpc == "default" ]; then vpc_id="default"; fi

    # Generate additional.tfvars with worker_groups_launch_template to pass module and create workers
            echo "worker_groups_launch_template = [" > additional.tfvars
    for (( i=0; i<${cluster_cloud_provisioner_node_group_count}; i++ ));
        do
            echo "{"  >> additional.tfvars
            add_tfvars "name" "\"$(eval echo '$'cluster_cloud_provisioner_node_group_${i}_name)\""
            add_tfvars "instance_type" "\"$(eval echo '$'cluster_cloud_provisioner_node_group_${i}_instance_type)\""
            add_tfvars "asg_desired_capacity" "$(eval echo '$'cluster_cloud_provisioner_node_group_${i}_asg_desired_capacity)"
            add_tfvars "asg_max_size" "$(eval echo '$'cluster_cloud_provisioner_node_group_${i}_asg_max_size)"
            add_tfvars "asg_min_size" "$(eval echo '$'cluster_cloud_provisioner_node_group_${i}_asg_min_size)"
            add_tfvars "root_volume_size" "$(eval echo '$'cluster_cloud_provisioner_node_group_${i}_root_volume_size)"
            add_tfvars "kubelet_extra_args" "\"$(eval echo '$'cluster_cloud_provisioner_node_group_${i}_kubelet_extra_args)\""
            add_tfvars "override_instance_types" "$(to_tf_list "$(eval echo '$'cluster_cloud_provisioner_node_group_${i}_override_instance_types)")"
            add_tfvars "spot_allocation_strategy" "\"$(eval echo '$'cluster_cloud_provisioner_node_group_${i}_spot_allocation_strategy)\""
            add_tfvars "spot_instance_pools" "$(eval echo '$'cluster_cloud_provisioner_node_group_${i}_spot_instance_pools)"
            add_tfvars "spot_max_price" "\"$(eval echo '$'cluster_cloud_provisioner_node_group_${i}_spot_max_price)\""
            add_tfvars "on_demand_base_capacity" "$(eval echo '$'cluster_cloud_provisioner_node_group_${i}_on_demand_base_capacity)"
            add_tfvars "on_demand_percentage_above_base_capacity" "$(eval echo '$'cluster_cloud_provisioner_node_group_${i}_on_demand_percentage_above_base_capacity)"
            if [ "$vpc_id" == "default" ]; then
                add_tfvars "public_ip" \""true"\"
            fi
            echo "},"  >> additional.tfvars
        done
            echo "]"  >> additional.tfvars

cat additional.tfvars
