# K8S

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| enable\_autoscaling | Enable / Disable Digital Ocean Kubernetes autoscale feature (e.g. `false`) | `bool` | `false` | no |
| max\_node\_count | Digital Ocean Kubernetes max nodes with autoscale feature (e.g. `3`) | `number` | `3` | no |
| min\_node\_count | Digital Ocean Kubernetes min nodes with autoscale feature (e.g. `1`) | `number` | `1` | no |
| name | (Required) Provide DigitalOcean cluster name | `string` | n/a | yes |
| node\_count | Digital Ocean Kubernetes node pool size (e.g. `2`) | `number` | `2` | no |
| node\_type | Digital Ocean Kubernetes default node pool type (e.g. `s-1vcpu-2gb` => 1vCPU, 2GB RAM) | `string` | `"s-1vcpu-2gb"` | no |
| region | (Required) Provide DigitalOcean region | `string` | n/a | yes |
| version | Provide DigitalOcean Kubernetes minor version (e.g. '1.16' or '1.15') | `string` | `"1.16"` | no |

## Outputs

No output.
