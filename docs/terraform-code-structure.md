## Filenames <!-- omit in toc -->

* [`init.tf`](#inittf)
* [`main.tf`](#maintf)
* [`outputs.tf`](#outputstf)
* [`vars.tf`](#varstf)
* [`README.md`](#readmemd)

Put the `.yaml`, `.sh` and other module-related files to subdir(s), `./templates` for instance.

Also, you can check our [style-guide](style-guide.md) for other best practices.

### `init.tf`

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

### `main.tf`

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

### `outputs.tf`

Sort alphabetically.

```terraform
output "_" {
  value       = _
  description = ""
}

```

### `vars.tf`

Sort alphabetically.

```terraform
variable "_" {
  type = _

  description = "_"
}

```

### `README.md`

```md
# Module name

Describe what the module does.

<!-- BEGINNING OF PRE-COMMIT-TERRAFORM DOCS HOOK -->
## Inputs

## Outputs

<!-- END OF PRE-COMMIT-TERRAFORM DOCS HOOK -->

```
