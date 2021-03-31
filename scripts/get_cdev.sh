#!/bin/bash
set -e
set -x

CDEV_VERSION=$(curl -s https://api.github.com/repos/shalb/cluster.dev/releases/latest | grep tag_name | cut -d '"' -f 4)

curl -Lo cdev.tgz https://github.com/shalb/cluster.dev/releases/download/${CDEV_VERSION}/cluster.dev_${CDEV_VERSION}_linux_amd64.tgz
tar -xzvf cdev.tgz -C /usr/local/bin
