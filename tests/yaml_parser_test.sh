#!/usr/bin/env bash

source ../bin/yaml.sh # provides parse_yaml and create_variables
source ../bin/logging.sh # PSR-3 compliant logging

CLUSTER_CONFIG_PATH='../.cluster.dev/minikube-one.yaml'

for CLUSTER_MANIFEST_FILE in $(find "$CLUSTER_CONFIG_PATH" -type f  || ERROR "Manifest file/folder can't be found" && exit 1 ); do

    yaml::parse "$CLUSTER_MANIFEST_FILE"
    yaml::create_variables "$CLUSTER_MANIFEST_FILE"
    yaml::check_that_required_variables_exist "$CLUSTER_CONFIG_PATH/$CLUSTER_MANIFEST_FILE"

echo "cluster apps to be installed:" ${cluster_apps[@]}
    for i in "${cluster_apps[@]}"; do echo "$i"; done
done
