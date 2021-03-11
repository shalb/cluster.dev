# Prerequisites

## Terraform

Cdev client uses the Terraform binary. The required Terraform version is ~13 or higher. Refer to the [Terraform installation instructions](https://www.terraform.io/downloads.html) to install Terraform.

Terraform installation example for linux amd64:

```bash
curl -O https://releases.hashicorp.com/terraform/0.14.7/terraform_0.14.7_linux_amd64.zip
unzip terraform_0.14.7_linux_amd64.zip
mv terraform /usr/local/bin/
```

## SOPS

For **creating** and **editing** SOPS secrets, cdev uses SOPS binary. But the SOPS binary is **not required** for decrypting and using SOPS secrets. As none of cdev reconcilation processes (build, plan, apply) requires SOPS to be performed, you don't have to install it for pipelines.

See [SOPS installation instructions](https://github.com/mozilla/sops#download) in official repo.

Also see [Secrets section](SOPS-secret) in this documentation.

