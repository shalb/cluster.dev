# K8S

<!-- BEGINNING OF PRE-COMMIT-TERRAFORM DOCS HOOK -->
## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|:----:|:-----:|:-----:|
| k8s\_version | Provide DigitalOcean Kubernetes minor version (e.g. '1.15' or higher) | string | `"1.17"` | no |
| max\_node\_count | Digital Ocean Kubernetes max nodes with autoscale feature (e.g. `2`) | number | `"2"` | no |
| min\_node\_count | Digital Ocean Kubernetes min nodes with autoscale feature (e.g. `1`) | number | `"1"` | no |
| name | (Required) Provide DigitalOcean cluster name | string | n/a | yes |
| node\_type | Digital Ocean Kubernetes default node pool type (e.g. `s-1vcpu-2gb` => 1vCPU, 2GB RAM) | string | `"s-1vcpu-2gb"` | no |
| region | (Required) Provide DigitalOcean region | string | n/a | yes |

## Outputs

| Name | Description |
|------|-------------|
| cluster\_endpoint |  |
| cluster\_status |  |

<!-- END OF PRE-COMMIT-TERRAFORM DOCS HOOK -->
