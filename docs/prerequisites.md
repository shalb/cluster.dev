# Prerequisites

This section describes preconditions and software necessary to start using cdev. <br> Please note that deploying to a certain cloud provider requires its own preconditions that are described in the [cloud providers prerequisites](#cloud-providers-prerequisites) subsection.

## cdev Prerequisites

To start using cdev please make sure that you have the following software installed:

* Linux amd64

* Git console client

* Bash

* Terraform

### Terraform

Cdev client uses the Terraform binary. The required Terraform version is ~13 or higher. Refer to the [Terraform installation instructions](https://www.terraform.io/downloads.html) to install Terraform.

Terraform installation example for Linux amd64:

```bash
curl -O https://releases.hashicorp.com/terraform/0.14.7/terraform_0.14.7_linux_amd64.zip
unzip terraform_0.14.7_linux_amd64.zip
mv terraform /usr/local/bin/
```

## Cloud providers prerequisites

### AWS prerequisites

1. Terraform version 13+.

2. AWS account.

3. AWS CLI installed.

4. kubectl installed.

5. [cdev installed](https://cluster.dev/installation/).

### DigitalOcean prerequisites

1. Terraform version 13+.

2. DigitalOcean account.

3. [doctl installed](https://docs.digitalocean.com/reference/doctl/how-to/install/).

4. [cdev installed](https://cluster.dev/installation/).
