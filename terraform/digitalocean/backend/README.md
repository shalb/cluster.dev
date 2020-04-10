## Backend

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| do\_token | Digital Ocean personal access token. | `string` | n/a | yes |
| name | (Required) Name for Spaces DO terraform backend | `string` | n/a | yes |
| region | (Required) Region for Spaces DO terraform backend | `string` | n/a | yes |
| spaces\_access\_id | Digital Ocean Spaces access id. | `string` | n/a | yes |
| spaces\_secret\_key | Digital Ocean Spaces secret key. | `string` | n/a | yes |

## Outputs

| Name | Description |
|------|-------------|
| spaces\_bucket\_name | Digital Ocean Spaces bucket name |
