# Terraform modules' structure

## Filenames <!-- omit in toc -->

* [`0-init.tf`](#0-inittf)
* [`1-main.tf`](#1-maintf)
* [`8-outputs.tf`](#8-outputstf)
* [`9-vars.tf`](#9-varstf)
* [`README.md`](#readmemd)

Put the `.yaml`, `.sh` and other module-related files directly to the module folder.

Also, look our [style-guide](style-guide.md) to check other best practices.

## `0-init.tf`

Sort `terraform_remote_state` alphabetically.

```terraform
#
# Init
#

terraform {
  required_version = ">= X.Y.Z" # minimum required Terraform version. Use the oldest version if possible.

  required_providers {
    _ = "~> X.Y" # version of the used provider, pin only MAJOR.MINOR. Use the newest version if possible.
  }

  backend "_" {}
}

provider "_" {}

#
# Remote states for import
#

data "terraform_remote_state" "" {}

```


## `1-main.tf`

Sort in logic order. Sort alphabetically when possible.

```terraform

#
# Get/transform exit data
#

locals {}

data "_" "_" {}

resource "_" "_" {}

module "_" {}

#
# Create resources
#

resource "_" "_" {}

module "_" {}

```

## `8-outputs.tf`

Sort alphabetically.

```terraform
output "_" {
  value       = _
  description = ""
}

```

## `9-vars.tf`

Sort alphabetically.

```terraform
variable "_" {
  type = _

  description = "_"
}

```

## `README.md`

```md
# Module name

Describe what the module does.

<!-- BEGINNING OF PRE-COMMIT-TERRAFORM DOCS HOOK -->
## Inputs

## Outputs

<!-- END OF PRE-COMMIT-TERRAFORM DOCS HOOK -->

```
