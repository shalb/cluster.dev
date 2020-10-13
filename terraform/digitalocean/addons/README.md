# Kubernetes Addons

Module which installs and destroys in AWS-based Kubernetes clusters:

ExternalDNS - using Helm chart
CertManager - using kubectl
Nginx-Ingress - using kubectl
ArgoCD - using Helm chart
Operator Lifecycle Manager - with default bash script
Keycloak Operator and Keycloak - using kubectl

<!-- BEGINNING OF PRE-COMMIT-TERRAFORM DOCS HOOK -->
## Requirements

| Name | Version |
|------|---------|
| terraform | ~> 0.12.20 |
| helm | ~> 1.0 |
| kubernetes | ~> 1.11 |
| null | ~> 2.1 |
| random | ~> 2.2 |
| template | ~> 2.1 |

## Providers

| Name | Version |
|------|---------|
| digitalocean | n/a |
| helm | ~> 1.0 |
| kubernetes | ~> 1.11 |
| null | ~> 2.1 |
| random | ~> 2.2 |
| template | ~> 2.1 |
| terraform | n/a |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| cluster\_cloud\_domain | Route 53 zone used as a domain restrictions for cert-manager and external-dns | `string` | `""` | no |
| cluster\_name | Full cluster name including user/organization | `string` | `""` | no |
| config\_path | path to a kubernetes config file | `string` | `"~/.kube/config"` | no |
| do\_token | Token required to access DO API | `string` | `""` | no |
| region | DO Region to apply for Addons configuration | `string` | n/a | yes |

## Outputs

| Name | Description |
|------|-------------|
| argocd\_pass | n/a |
| argocd\_url | n/a |
| argocd\_user | n/a |

<!-- END OF PRE-COMMIT-TERRAFORM DOCS HOOK -->
