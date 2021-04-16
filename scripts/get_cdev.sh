#!/bin/bash
set -e
set -x
BIN_DIR="${HOME}/.cluster.dev/bin/"
CDEV_LATEST_VERSION=$(curl -s https://api.github.com/repos/shalb/cluster.dev/releases/latest | grep tag_name | cut -d '"' -f 4)

PROFILE_FILE=""


TMP_DIR=$(mktemp -d)

OS=""
case $(uname) in
    "Linux") OS="linux";;
    "Darwin") OS="darwin";;
    *)
        echo "Unsupported os type $(uname)"
        exit 1
        ;;
esac

ARCH=""
case $(uname -m) in
    "x86_64") ARCH="amd64";;
    "arm64") ARCH="arm64";;
    *)
        echo "Unsupported arch $(uname -m)"
        exit 1
        ;;
esac

if [ "$(uname)" != "Darwin" ]; then
    if [ -e "${HOME}/.bashrc" ]; then
        PROFILE_FILE="${HOME}/.bashrc"
    elif [ -e "${HOME}/.bash_profile" ]; then
        PROFILE_FILE="${HOME}/.bash_profile"
    fi
else
    if [ -e "${HOME}/.bash_profile" ]; then
        PROFILE_FILE="${HOME}/.bash_profile"
    fi
    if [ -e "${HOME}/.bashrc" ]; then
        PROFILE_FILE="${HOME}/.bashrc"
    fi
    if [ -e "${HOME}/.zshrc" ]; then
        PROFILE_FILE="${HOME}/.zshrc"
    fi
fi

mkdir -p ${BIN_DIR}

curl -fLo ${TMP_DIR}/cdev.tgz https://github.com/shalb/cluster.dev/releases/download/${CDEV_LATEST_VERSION}/cdev-${CDEV_LATEST_VERSION}-${OS}-${ARCH}.tar.gz
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
