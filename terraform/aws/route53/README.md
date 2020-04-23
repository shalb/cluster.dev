# Route 53

Creates a domain zone in format `cluster_fullname`.`cluster\_domain`
and returns zone_id and name_servers.

<!-- BEGINNING OF PRE-COMMIT-TERRAFORM DOCS HOOK -->
## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|:----:|:-----:|:-----:|
| cluster\_domain | Default domain for cluster records | string | n/a | yes |
| cluster\_fullname | Full name of the cluster | string | n/a | yes |
| region | The AWS region. | string | n/a | yes |
| zone_delegation | If true - a NS records in cluster_domain(cluster.dev) to be created by external scripts. | boolean | false | yes |

## Outputs

| Name | Description | Type | Default | Required |
|------|-------------|:----:|:-----:|:-----:|
| zone\_id | Domain ZoneID | string | n/a | yes |
| name\_servers | NS records for new zone | list | n/a | yes |
<!-- END OF PRE-COMMIT-TERRAFORM DOCS HOOK -->
