#!/usr/bin/env bash
# Script emulates GitHub Action execution in local env.
# Used for the testing cluster creation and performs basic tests

# Import variables
readonly SRC_PATH=$(realpath $(dirname $(readlink -f $0))/../)

. config.sh
cd "${SRC_PATH}"
source "${SRC_PATH}/bin/logging.sh"

readonly GIT_SHORT_COMMIT="$(git rev-parse --short HEAD)"
readonly DOCKER_IMAGE_NAME="cluster.dev:${GIT_SHORT_COMMIT}-local-tests"

# Delete .terraform dirs.
docker run --rm --workdir /github/workspace --rm -v "${SRC_PATH}:/github/workspace" alpine find ./ -name .terraform -type d -exec rm -rf {} +

# Build with image --no-cache (always build new).
docker build --no-cache -t "${DOCKER_IMAGE_NAME}" .

# Get from config.sh
readonly USER="${AWS_ACCESS_KEY_ID}"
readonly PASS="${AWS_SECRET_ACCESS_KEY}"
readonly CONFIG_PATH="${CLUSTER_CONFIG_PATH}"
readonly TIMEOUT="${ACTION_TIMEOUT}"

# Trap ctrl+c to remove docker container and kill timeout script.
trap ctrl_c INT
function ctrl_c {
    docker rm -f "clusterdev-test-${GIT_SHORT_COMMIT}"
    kill "${timer_pid}"
}

# Script waits for $1 seconds and than remove $2 containet.
${SRC_PATH}/tests/timeout.sh "${TIMEOUT}" "clusterdev-test-${GIT_SHORT_COMMIT}" &
timer_pid=$!

# Run docker in localhost
docker run  -d --rm \
            --name "clusterdev-test-${GIT_SHORT_COMMIT}" \
            --workdir /tests/workspace \
            -v "${SRC_PATH}:/tests/workspace" \
            -e GIT_PROVIDER="test-run" \
            -e GIT_REPO_NAME="test-run" \
            -e GIT_REPO_ROOT="/tests/workspace" \
            "${DOCKER_IMAGE_NAME}" "${CONFIG_PATH}" "${USER}" "${PASS}"

sleep 1

# Show pipeline containet output.
docker logs -f "clusterdev-test-${GIT_SHORT_COMMIT}"
