# Minikune

Creates a single instance Kubernetes installation with kubeadm.


<!-- BEGINNING OF PRE-COMMIT-TERRAFORM DOCS HOOK -->
## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|:----:|:-----:|:-----:|
| aws\_instance\_type | Instance size | string | n/a | yes |
| cluster\_name | Name of the cluster | string | n/a | yes |
| hosted\_zone | DNS zone to use in cluster | string | n/a | yes |
| region | The AWS region. | string | n/a | yes |

<!-- END OF PRE-COMMIT-TERRAFORM DOCS HOOK -->
