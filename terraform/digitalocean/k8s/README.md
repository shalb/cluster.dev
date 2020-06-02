# K8S

<!-- BEGINNING OF PRE-COMMIT-TERRAFORM DOCS HOOK -->
## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| max\_node\_count | Digital Ocean Kubernetes max nodes with autoscale feature (e.g. `2`) | `number` | `3` | no |
| min\_node\_count | Digital Ocean Kubernetes min nodes with autoscale feature (e.g. `1`) | `number` | `1` | no |
| name | (Required) Provide DigitalOcean cluster name | `string` | n/a | yes |
| node\_type | Digital Ocean Kubernetes default node pool type (e.g. `s-1vcpu-2gb` => 1vCPU, 2GB RAM) | `string` | `"s-1vcpu-2gb"` | no |
| region | (Required) Provide DigitalOcean region | `string` | n/a | yes |
| k8s_version | Provide DigitalOcean Kubernetes minor version (e.g. '1.16' or '1.15') | `string` | `"1.16"` | no |

## Outputs

| Name | Description |
|------|-------------|
| cluster\_status | A string indicating the current status of the cluster. Potential values include running, provisioning, and errored. |
| cluster\_endpoint | The base URL of the API server on the Kubernetes master node. |

<!-- END OF PRE-COMMIT-TERRAFORM DOCS HOOK -->
