# EKS module

Module creates an EKS cluster inside AWS.
To use it need to generate `worker_groups_launch_template` inside a `worker_groups.tfvars`
See example file in `worker_groups.tfvars.example`

<!-- BEGINNING OF PRE-COMMIT-TERRAFORM DOCS HOOK -->
## Requirements

| Name | Version |
|------|---------|
| terraform | ~> 0.12.20 |
| aws | ~> 2.64.0 |
| kubernetes | ~> 1.11.3 |
| local | ~> 1.4.0 |
| null | ~> 2.1.2 |
| random | ~> 2.2.1 |

## Providers

| Name | Version |
|------|---------|
| aws | ~> 2.64.0 |
| terraform | n/a |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| availability\_zones | The AWS Availability Zone(s) inside region. | `list` | n/a | yes |
| cluster\_name | Name of the EKS cluster. Also used as a prefix in names of related resources. | `string` | n/a | yes |
| cluster\_version | Kubernetes version to use for the EKS cluster. | `string` | `"1.16"` | no |
| region | The AWS region. | `string` | n/a | yes |
| tags | A map of tags to add to all resources. | `map(string)` | `{}` | no |
| vpc\_id | VPC where the cluster and workers will be deployed. | `string` | n/a | yes |
| worker\_additional\_security\_group\_ids | A list of additional security group ids to attach to worker instances | `list(string)` | `[]` | no |
| worker\_groups\_launch\_template | A list of maps defining worker group configurations to be defined using AWS Launch Templates. See workers\_group\_defaults for valid keys. | `any` | `[]` | no |
| workers\_subnets\_type | Type of subnets to use on worker nodes: public or private | `string` | `"private"` | no |

## Outputs

| Name | Description |
|------|-------------|
| cluster\_id | n/a |
| cluster\_oidc\_issuer\_url | n/a |

<!-- END OF PRE-COMMIT-TERRAFORM DOCS HOOK -->
