#!/bin/bash

sleep 1
ACTION_TIMEOUT=$1
CONTAINER_NAME=$2

  for seconds in $(seq 1 ${ACTION_TIMEOUT}); do
    if [ -z  "$(docker ps| grep ${CONTAINER_NAME})" ]; then
       exit 0
    fi
    sleep 1
  done
  echo "Pipeline executing more than ${seconds}s. Stoped by timeout"
  docker rm -f ${CONTAINER_NAME}
