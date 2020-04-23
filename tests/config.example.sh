#!/usr/bin/env bash

export AWS_ACCESS_KEY_ID="000000000000000000000" # aws_secret_access_key
export AWS_SECRET_ACCESS_KEY="00000000000000000000000000000000000000" # aws_access_key_id
export CLUSTER_CONFIG_PATH=".cluster.dev/aws-minikube.yaml" # Relative path from project root
export ACTION_TIMEOUT="600" # Timeout in seconds. After timeout action will be terminated container will be stopped.
