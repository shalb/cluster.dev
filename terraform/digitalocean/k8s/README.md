# K8S

<!-- BEGINNING OF PRE-COMMIT-TERRAFORM DOCS HOOK -->
## Requirements

| Name | Version |
|------|---------|
| terraform | ~> 0.12.0 |
| digitalocean | ~> 1.18.0 |
| kubernetes | ~> 1.11.0 |

## Providers

| Name | Version |
|------|---------|
| digitalocean | ~> 1.18.0 |
| local | n/a |
| terraform | n/a |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| cluster\_name | (Required) Provide DigitalOcean cluster name | `string` | n/a | yes |
| config\_output\_path | Where to save the Kubectl config file (if `write_kubeconfig = true`). Assumed to be a directory if the value ends with a forward slash `/`. | `string` | `"./"` | no |
| k8s\_version | Provide DigitalOcean Kubernetes minor version (e.g. '1.15' or higher) | `string` | `"1.17"` | no |
| max\_node\_count | Digital Ocean Kubernetes max nodes with autoscale feature (e.g. `2`) | `number` | `2` | no |
| min\_node\_count | Digital Ocean Kubernetes min nodes with autoscale feature (e.g. `1`) | `number` | `1` | no |
| node\_type | Digital Ocean Kubernetes default node pool type (e.g. `s-1vcpu-2gb` => 1vCPU, 2GB RAM) | `string` | `"s-1vcpu-2gb"` | no |
| region | (Required) Provide DigitalOcean region | `string` | n/a | yes |
| write\_kubeconfig | Whether to write a Kubectl config file containing the cluster configuration. Saved to `config_output_path`. | `bool` | `true` | no |

## Outputs

| Name | Description |
|------|-------------|
| cluster\_endpoint | n/a |
| cluster\_status | n/a |
| kubernetes\_config | n/a |

<!-- END OF PRE-COMMIT-TERRAFORM DOCS HOOK -->
