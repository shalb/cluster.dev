# Terraform modules structure

## Filenames <!-- omit in toc -->

* [`0-init.tf`](#0-inittf)
* [`1-main.tf`](#1-maintf)
* [`8-outputs.tf`](#8-outputstf)
* [`9-vars.tf`](#9-varstf)
* [`README.md`](#readmemd)

`.yaml`, `.sh` and other related to module files put directly to module folder.

Also, look our [style-guide](style-guide.md) for check other best practices.

## `0-init.tf`

Sort `terraform_remote_state` alphabetically.

```terraform
#
# Init
#

terraform {
  required_version = ">= X.Y.Z" # minimum required terraform version. Older - better.

  required_providers {
    _ = "~> X.Y" # used provider version, pin only MAJOR.MINOR. Newest - better.
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

In alphabet order

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

Describe What this module do.

<!-- BEGINNING OF PRE-COMMIT-TERRAFORM DOCS HOOK -->
## Inputs

## Outputs

<!-- END OF PRE-COMMIT-TERRAFORM DOCS HOOK -->

```
