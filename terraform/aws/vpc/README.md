# VPC

Module creates or re-use existing VPC.

If vpc_id equal to:
  "create" - creates VPC by terraform-aws-vpc module
  "default" - use default VPC in provided region
  "vpc_id"  - it would use provided id for existing VPC but networks should be tagged
  with "cluster.dev/subnet_type" = "private|public" tags to become visible for module.


<!-- BEGINNING OF PRE-COMMIT-TERRAFORM DOCS HOOK -->
## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|:----:|:-----:|:-----:|
| availability\_zones | The AWS Availability Zone(s) inside region. | list | n/a | yes |
| cluster\_name | Name of the cluster. | string | n/a | yes |
| region | The AWS region. | string | n/a | yes |
| vpc\_cidr | Vpc CIDR | string | `"10.8.0.0/18"` | no |
| vpc\_id | Vpc ID, or create or default | string | `"default"` | no |

## Outputs

| Name | Description |
|------|-------------|
| private\_subnets |  |
| public\_subnets |  |
| vpc\_cidr |  |
| vpc\_id |  |

<!-- END OF PRE-COMMIT-TERRAFORM DOCS HOOK -->
