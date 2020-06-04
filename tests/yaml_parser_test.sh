#!/usr/bin/env bash

source ../bin/common.sh

    CLUSTER_MANIFEST_FILE='../.cluster.dev/aws-eks.yaml'
    ~/yamltoenv -f "$CLUSTER_MANIFEST_FILE"
    source <(~/yamltoenv -f "$CLUSTER_MANIFEST_FILE")


echo "---------------"

            # Generate worker_groups.tfvars with worker_groups_launch_template to pass module and create workers
                    echo "worker_groups_launch_template = [" > worker_groups.tfvars
            for (( i=0; i<${cluster_cloud_provisioner_node_group_count}; i++ ));
                do
                    echo "{"
                    eval name='$'cluster_cloud_provisioner_node_group_${i}_name; if [ ! -z $name ]; then echo name =  \""$name"\" ; fi
                    eval instance_type='$'cluster_cloud_provisioner_node_group_${i}_instance_type; if [ ! -z $instance_type ]; then echo instance_type =  \""$instance_type"\" ; fi
                    eval asg_desired_capacity='$'cluster_cloud_provisioner_node_group_${i}_asg_desired_capacity; if [ ! -z $asg_desired_capacity ]; then echo asg_desired_capacity =  "$asg_desired_capacity" ; fi
                    eval asg_max_size='$'cluster_cloud_provisioner_node_group_${i}_asg_max_size; if [ ! -z $asg_max_size ]; then echo asg_max_size =  "$asg_max_size" ; fi
                    eval asg_min_size='$'cluster_cloud_provisioner_node_group_${i}_asg_min_size; if [ ! -z $asg_min_size ]; then echo asg_min_size =  "$asg_min_size" ; fi
                    eval root_volume_size='$'cluster_cloud_provisioner_node_group_${i}_root_volume_size; if [ ! -z $root_volume_size ]; then echo root_volume_size =  "$root_volume_size" ; fi
                    eval kubelet_extra_args='$'cluster_cloud_provisioner_node_group_${i}_kubelet_extra_args; if [ ! -z $kubelet_extra_args ]; then echo kubelet_extra_args =  \""$kubelet_extra_args"\" ; fi
                    eval instance_type_override='$'cluster_cloud_provisioner_node_group_${i}_instance_type_override; if [ ! -z "$instance_type_override" ]; then echo instance_type_override = $(to_tf_list "$instance_type_override") ; fi
                    eval spot_allocation_strategy='$'cluster_cloud_provisioner_node_group_${i}_spot_allocation_strategy; if [ ! -z $spot_allocation_strategy ]; then echo spot_allocation_strategy =  \""$spot_allocation_strategy"\" ; fi
                    eval spot_instance_pools='$'cluster_cloud_provisioner_node_group_${i}_spot_instance_pools; if [ ! -z $spot_instance_pools ]; then echo spot_instance_pools =  "$spot_instance_pools" ; fi
                    eval spot_max_price='$'cluster_cloud_provisioner_node_group_${i}_spot_max_price; if [ ! -z $spot_max_price ]; then echo spot_max_price =  \""$spot_max_price"\" ; fi
                    eval on_demand_base_capacity='$'cluster_cloud_provisioner_node_group_${i}_on_demand_base_capacity; if [ ! -z $on_demand_base_capacity ]; then echo on_demand_base_capacity =  "$on_demand_base_capacity" ; fi
                    eval on_demand_percentage_above_base_capacity='$'cluster_cloud_provisioner_node_group_${i}_on_demand_percentage_above_base_capacity; if [ ! -z $on_demand_percentage_above_base_capacity ]; then echo on_demand_percentage_above_base_capacity =  "$on_demand_percentage_above_base_capacity" ; fi
                    if [ "$vpc_id" == "default" ]; then
                        echo public_ip =  \""true"\"
                    fi
                    # TODO: Add bastion access `key_name`
                    # eval name='$'cluster_cloud_provisioner_node_group_${i}_key_name; if [ ! -z $key_name ]; then echo key_name =  \""ec2-key"\" ; fi
                    echo "},"
                done
                    echo "]"
