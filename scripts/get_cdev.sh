#!/bin/bash
set -e
set -x
BIN_DIR="${HOME}/.cluster.dev/bin/"
CDEV_LATEST_VERSION=$(curl -s https://api.github.com/repos/shalb/cluster.dev/releases/latest | grep tag_name | cut -d '"' -f 4)

PROFILE_FILE=""


TMP_DIR=$(mktemp -d)

if [ -e "${HOME}/.bash_profile" ]; then
    PROFILE_FILE="${HOME}/.bash_profile"
elif [ -e "${HOME}/.bashrc" ]; then
    PROFILE_FILE="${HOME}/.bashrc"
fi

mkdir -p ${BIN_DIR}

curl -fLo ${TMP_DIR}/cdev.tgz https://github.com/shalb/cluster.dev/releases/download/${CDEV_LATEST_VERSION}/cluster.dev_${CDEV_LATEST_VERSION}_linux_amd64.tgz
tar -xzvf ${TMP_DIR}/cdev.tgz -C ${BIN_DIR}

if [ -n "${PROFILE_FILE}" ]; then
    ADD_CDEV_LINE="export PATH=\$PATH:\$HOME/.cluster.dev/bin"
    if ! grep -q "# add cdev to the PATH" "${PROFILE_FILE}"; then
        printf "\\n# add cdev to the PATH\\n%s\\n" "${ADD_CDEV_LINE}" >> "${PROFILE_FILE}"
    fi
    
    echo "Please restart your shell or add $HOME/.cluster.dev/bin to your \$PATH"
else
    echo "Please add $HOME/.cluster.dev/bin to your \$PATH"
fi

rm -rf ${TMP_DIR}