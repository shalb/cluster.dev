## Backend

Module creates a state storage bucket in DigitalOcean Spaces.

<!-- BEGINNING OF PRE-COMMIT-TERRAFORM DOCS HOOK -->
## Requirements

| Name | Version |
|------|---------|
| terraform | ~> 0.12.0 |
| digitalocean | ~> 1.18.0 |

## Providers

| Name | Version |
|------|---------|
| digitalocean | ~> 1.18.0 |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| do\_spaces\_backend\_bucket | (Required) Name for Spaces DO terraform backend | `string` | n/a | yes |
| region | (Required) Region for Spaces DO terraform backend | `string` | n/a | yes |

## Outputs

| Name | Description |
|------|-------------|
| spaces\_bucket\_name | Digital Ocean Spaces bucket name |

<!-- END OF PRE-COMMIT-TERRAFORM DOCS HOOK -->
