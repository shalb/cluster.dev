#!/bin/bash

readonly SRC_PATH=$(realpath $(dirname $(readlink -f $0))/../)
cd ${SRC_PATH}

readonly GIT_SHOT_COMMIT=$(git rev-parse --short HEAD)
readonly DOCKER_IMAGE_NAME="cluster.dev:${GIT_SHOT_COMMIT}-local-tests"

docker build -t ${DOCKER_IMAGE_NAME} .

readonly USER="AWS_SECRET_KEY"
readonly PASS="AWS_SECRET_TOKEN"

docker run --name NAME --workdir /github/workspace --rm -v "${SRC_PATH}":"/github/workspace" \
           -e GITHUB_REPOSITORY="shalb" \
           ${DOCKER_IMAGE_NAME}  "./.cluster.dev/gh-5.yaml" "${USER}" "${PASS}"
