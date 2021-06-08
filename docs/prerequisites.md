# cdev Prerequisites

Supported operation systems:

* Linux amd64

* Darwin amd64

To start using cdev please make sure that you have the following software installed:

* Git console client

* Terraform

## Terraform

cdev client uses the Terraform binary. The required Terraform version is ~13 or higher. Refer to the [Terraform installation instructions](https://www.terraform.io/downloads.html) to install Terraform.

Terraform installation example for Linux amd64:

```bash
curl -O https://releases.hashicorp.com/terraform/0.14.7/terraform_0.14.7_linux_amd64.zip
unzip terraform_0.14.7_linux_amd64.zip
mv terraform /usr/local/bin/
```
