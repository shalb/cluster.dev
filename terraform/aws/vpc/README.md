# VPC

Module creates or re-use existing VPC.

If vpc_id equal to:
  "create" - creates VPC by terraform-aws-vpc module
  "default" - use default VPC in provided region
  "vpc_id"  - it would use provided id for existing VPC but networks should be tagged
  with "cluster.dev/subnet_type" = "private|public" tags to become visible for module.


<!-- BEGINNING OF PRE-COMMIT-TERRAFORM DOCS HOOK -->
## Requirements

| Name | Version |
|------|---------|
| terraform | ~> 0.12.20 |
| aws | ~> 2.64.0 |
| null | ~> 2.1 |

## Providers

| Name | Version |
|------|---------|
| aws | ~> 2.64.0 |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| availability\_zones | The AWS Availability Zone(s) inside region. | `list` | n/a | yes |
| cluster\_name | Name of the cluster. | `string` | n/a | yes |
| region | The AWS region. | `string` | n/a | yes |
| vpc\_cidr | Vpc CIDR | `string` | `"10.8.0.0/18"` | no |
| vpc\_id | Vpc ID, or create or default | `string` | `"default"` | no |

## Outputs

| Name | Description |
|------|-------------|
| private\_subnets | n/a |
| public\_subnets | n/a |
| vpc\_cidr | n/a |
| vpc\_id | If vpc\_id = "create" - dispay vpc\_id of created by module "default" - displey vpc\_id of default vpc in the provided region "vpc\_id"  - existing vpc - display just provided id |

<!-- END OF PRE-COMMIT-TERRAFORM DOCS HOOK -->
