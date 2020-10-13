# Route 53

Creates a domain zone in format `cluster_fullname`.`cluster\_domain`
and returns zone_id and name_servers.

<!-- BEGINNING OF PRE-COMMIT-TERRAFORM DOCS HOOK -->
## Requirements

| Name | Version |
|------|---------|
| terraform | ~> 0.12.0 |
| aws | ~> 2.23 |

## Providers

| Name | Version |
|------|---------|
| aws | ~> 2.23 |
| null | n/a |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| cluster\_domain | Default domain for cluster records | `string` | n/a | yes |
| cluster\_name | Full name of the cluster | `string` | n/a | yes |
| dns\_manager\_url | Endpoint to create a default zone in cluster.dev domain | `string` | `"https://usgrtk5fqj.execute-api.eu-central-1.amazonaws.com/prod"` | no |
| email | email of user which requests a default zone | `string` | `"domain-request@cluster.dev"` | no |
| region | The AWS region. | `string` | n/a | yes |
| zone\_delegation | If true - a NS records in cluster\_domain(cluster.dev) to be created by external scripts | `string` | `false` | no |

## Outputs

| Name | Description |
|------|-------------|
| name\_servers | n/a |
| zone\_id | n/a |

<!-- END OF PRE-COMMIT-TERRAFORM DOCS HOOK -->
