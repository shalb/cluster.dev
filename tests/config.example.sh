#!/usr/bin/env bash

AWS_ACCESS_KEY_ID=""
AWS_SECRET_ACCESS_KEY=""

ACTION_TIMEOUT="600" # Timeout in seconds. After timeout action will be terminated container will be stopped.
GH_ACTION_WORKFLOW_PATH="./.cluster.dev/gh-7.yaml" # Relative path from project root
