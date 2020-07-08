# Project

<!-- BEGINNING OF PRE-COMMIT-TERRAFORM DOCS HOOK -->
## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|:----:|:-----:|:-----:|
| description | A description for this project | string | `""` | no |
| environment | (optional) Edit values below for every environment name that you want | map | `{ "dev": "Development", "prod": "Production", "stage": "Staging" }` | no |
| name | (Required) Name for your project at Digital Ocean | string | n/a | yes |
| purpose | The purpose for your project. (for example: k8s or any), default - Web Application | string | `"Web Application"` | no |
| resources | (Optional) List with all resources to be part of this project | list | `[]` | no |

## Outputs

| Name | Description |
|------|-------------|
| created\_at | The date and time when the project was created, (ISO8601) |
| id | The id of the project |
| owner\_id | The id of the project owner |
| owner\_uuid | The unique universal identifier of the project owner |
| updated\_at | The date and time when the project was last updated, (ISO8601) |

<!-- END OF PRE-COMMIT-TERRAFORM DOCS HOOK -->
