# VPC

<!-- BEGINNING OF PRE-COMMIT-TERRAFORM DOCS HOOK -->
## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|:----:|:-----:|:-----:|
| availability\_zones | The AWS Availability Zone(s) inside region. | list | n/a | yes |
| cluster\_name | Name of the cluster. | string | n/a | yes |
| region | The AWS region. | string | n/a | yes |
| vpc\_cidr | Vpc CIDR | string | `"10.8.0.0/18"` | no |

## Outputs

| Name | Description |
|------|-------------|
| vpc\_id |  |

<!-- END OF PRE-COMMIT-TERRAFORM DOCS HOOK -->
