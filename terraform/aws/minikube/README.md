# Minikune

Creates a single instance Kubernetes installation with kubeadm.


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
| template | n/a |
| terraform | n/a |
| tls | n/a |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| aws\_instance\_type | Instance size | `string` | n/a | yes |
| cluster\_name | Name of the cluster | `string` | n/a | yes |
| hosted\_zone | DNS zone to use in cluster | `string` | n/a | yes |
| region | The AWS region. | `string` | n/a | yes |

## Outputs

No output.

<!-- END OF PRE-COMMIT-TERRAFORM DOCS HOOK -->
