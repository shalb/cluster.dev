# VPC

Module creates or re-use existing VPC in DigitalOcean.

If vpc_id =
  "create" - creates VPC
  "default" - use default VPC in provided region
  "vpc_id"  - it would use provided id
}

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
| cluster\_name | Name of the cluster. | `string` | n/a | yes |
| ip\_range | The range of IP addresses for the VPC in CIDR notation | `string` | `"10.8.0.0/18"` | no |
| region | The DigitalOcean region. | `string` | n/a | yes |
| vpc\_id | Vpc ID, or create or default | `string` | `"create"` | no |

## Outputs

| Name | Description |
|------|-------------|
| ip\_range | n/a |
| vpc\_id | If vpc\_id = "create" - dispay vpc\_id of created by module "default" - displey vpc\_id of default vpc in the provided region "vpc\_id"  - existing vpc - display just provided id |

<!-- END OF PRE-COMMIT-TERRAFORM DOCS HOOK -->
