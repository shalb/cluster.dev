#!/usr/bin/env bash
# Script emulates GitHub Action execution in local env.
# Used for the testing cluster creation and performs basic tests

# Import variables
. config.sh

readonly SRC_PATH=$(realpath $(dirname $(readlink -f $0))/../)
cd ${SRC_PATH}

readonly GIT_SHORT_COMMIT=$(git rev-parse --short HEAD)
readonly DOCKER_IMAGE_NAME="cluster.dev:${GIT_SHORT_COMMIT}-local-tests"

docker build -t ${DOCKER_IMAGE_NAME} .

# Get from config.sh
readonly USER="${AWS_ACCESS_KEY_ID}"
readonly PASS="${AWS_SECRET_ACCESS_KEY}"
readonly WORKFLOW_PATH="${GH_ACTION_WORKFLOW_PATH}"

# Run docker in localhost
docker run --name clusterdev-test-GIT_SHORT_COMMIT --workdir /github/workspace --rm -v "${SRC_PATH}":"/github/workspace" \
           -e GITHUB_REPOSITORY="shalb" \
           ${DOCKER_IMAGE_NAME} "${WORKFLOW_PATH}" "${USER}" "${PASS}"
