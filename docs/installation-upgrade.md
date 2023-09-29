# Installation and Upgrade

## Prerequisites

To start using Cluster.dev please make sure that you comply with the following preconditions. 

Supported operating systems:

* Linux amd64

* Darwin amd64

Required software installed:

* Git console client

* Terraform

### Terraform

The Cluster.dev client uses the Terraform binary. The required Terraform version is 1.4 or higher. Refer to the [Terraform installation instructions](https://www.terraform.io/downloads.html) to install Terraform.

## Install From Script

!!! tip

    This is the easiest way to have the Cluster.dev client installed. For other options see the [Install From Sources](#install-from-sources) section.

Cluster.dev has an installer script that takes the latest version of Cluster.dev client and installs it for you locally.<br> 

Fetch the script and execute it locally with the command:

```bash
curl -fsSL https://raw.githubusercontent.com/shalb/cluster.dev/master/scripts/get_cdev.sh | sh
```

## Install From Sources

### Download from release

Each stable version of Cluster.dev has a binary that can be downloaded and installed manually. The documentation is suitable for **v0.4.0 or higher** of the Cluster.dev client.

Installation example for Linux amd64:

1. Download your desired version from the [releases page](https://github.com/shalb/cluster.dev/releases).

2. Unpack it.

3. Find the Cluster.dev binary in the unpacked directory.

4. Move the binary to the bin folder (/usr/local/bin/).

### Building from source

Go version 16 or higher is required - see [Golang installation instructions](https://golang.org/doc/install).

To build the Cluster.dev client from source:

1. Clone the Cluster.dev Git repo:

     ```bash
     git clone https://github.com/shalb/cluster.dev/
     ```

2. Build the binary:

     ```bash
     cd cluster.dev/ && make
     ```

3. Check Cluster.dev and move the binary to the bin folder:

     ```bash
     ./bin/cdev --help
     mv ./bin/cdev /usr/local/bin/
     ```