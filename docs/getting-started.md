# Getting Started

## Cdev Prerequisites

Supported operation systems:

* Linux amd64

* Darwin amd64

To start using cdev please make sure that you have the following software installed:

* Git console client

* Terraform

### Terraform

Cdev client uses the Terraform binary. The required Terraform version is ~13 or higher. Refer to the [Terraform installation instructions](https://www.terraform.io/downloads.html) to install Terraform.

Terraform installation example for Linux amd64:

```bash
curl -O https://releases.hashicorp.com/terraform/0.14.7/terraform_0.14.7_linux_amd64.zip
unzip terraform_0.14.7_linux_amd64.zip
mv terraform /usr/local/bin/
```

## Cdev Install

### From script *<small>recommended</small>*

Cdev has an installer script that takes the latest version of cdev and installs it for you locally.<br> 

You can fetch the script and execute it locally:

```bash
curl -fsSL https://raw.githubusercontent.com/shalb/cluster.dev/master/scripts/get_cdev.sh | bash
```

!!! tip

    We recommend installation from script as the easiest and the quickest way to have cdev installed. For other options, please see [Cdev Install Reference](https://cluster.dev/cdev-installation-reference/) section.

