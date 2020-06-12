# Kubernetes Addons

Module which installs and destroys in AWS-based Kubernetes clusters:

ExternalDNS - using Helm chart
CertManager - using kubectl
Nginx-Ingress - using kubectl
ArgoCD - using Helm chart

<!-- BEGINNING OF PRE-COMMIT-TERRAFORM DOCS HOOK -->
## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|:----:|:-----:|:-----:|
| cluster\_cloud\_domain | Route 53 zone used as a domain restrictions for cert-manager and external-dns | string | `""` | no |
| cluster\_name | Full cluster name including user/organization | string | `""` | no |
| config\_path | path to a kubernetes config file | string | `"~/.kube/config"` | no |
| eks | Define if addons would be deployed to EKS cluster | bool | `"false"` | no |
| region | AWS Region to apply for Addons configuration | string | n/a | yes |

## Outputs

| Name | Description |
|------|-------------|
| argocd\_pass |  |
| argocd\_url |  |
| argocd\_user |  |

<!-- END OF PRE-COMMIT-TERRAFORM DOCS HOOK -->
