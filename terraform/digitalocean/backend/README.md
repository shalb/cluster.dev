## Backend

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| do\_token | Digital Ocean personal access token. | `string` | n/a | yes |
| do_spaces_backend_bucket | (Required) Name for Spaces DO terraform backend | `string` | n/a | yes |
| region | (Required) Region for Spaces DO terraform backend | `string` | n/a | yes |

## Outputs

| Name | Description |
|------|-------------|
| spaces\_bucket\_name | Digital Ocean Spaces bucket name |
