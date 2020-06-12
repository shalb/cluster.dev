# EKS module

Module creates an EKS cluster inside AWS.
To use it need to generate `worker_groups_launch_template` inside a `worker_groups.tfvars`
See example file in `worker_groups.tfvars.example`

<!-- BEGINNING OF PRE-COMMIT-TERRAFORM DOCS HOOK -->
## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|:----:|:-----:|:-----:|
| attach\_worker\_cni\_policy | Whether to attach the Amazon managed `AmazonEKS_CNI_Policy` IAM policy to the default worker IAM role. WARNING: If set `false` the permissions must be assigned to the `aws-node` DaemonSet pods via another method or nodes will not be able to join the cluster. | bool | `"true"` | no |
| availability\_zones | The AWS Availability Zone(s) inside region. | list | n/a | yes |
| cluster\_create\_security\_group | Whether to create a security group for the cluster or attach the cluster to `cluster_security_group_id`. | bool | `"true"` | no |
| cluster\_create\_timeout | Timeout value when creating the EKS cluster. | string | `"30m"` | no |
| cluster\_delete\_timeout | Timeout value when deleting the EKS cluster. | string | `"15m"` | no |
| cluster\_enabled\_log\_types | A list of the desired control plane logging to enable. For more information, see Amazon EKS Control Plane Logging documentation (https://docs.aws.amazon.com/eks/latest/userguide/control-plane-logs.html) | list(string) | `[]` | no |
| cluster\_encryption\_config | Configuration block with encryption configuration for the cluster. See examples/secrets_encryption/main.tf for example format | object | `[]` | no |
| cluster\_endpoint\_private\_access | Indicates whether or not the Amazon EKS private API server endpoint is enabled. | bool | `"false"` | no |
| cluster\_endpoint\_private\_access\_cidrs | List of CIDR blocks which can access the Amazon EKS private API server endpoint, when public access is disabled | list(string) | `[ "0.0.0.0/0" ]` | no |
| cluster\_endpoint\_public\_access | Indicates whether or not the Amazon EKS public API server endpoint is enabled. | bool | `"true"` | no |
| cluster\_endpoint\_public\_access\_cidrs | List of CIDR blocks which can access the Amazon EKS public API server endpoint. | list(string) | `[ "0.0.0.0/0" ]` | no |
| cluster\_iam\_role\_name | IAM role name for the cluster. Only applicable if manage_cluster_iam_resources is set to false. | string | `""` | no |
| cluster\_log\_kms\_key\_id | If a KMS Key ARN is set, this key will be used to encrypt the corresponding log group. Please be sure that the KMS Key has an appropriate key policy (https://docs.aws.amazon.com/AmazonCloudWatch/latest/logs/encrypt-log-data-kms.html) | string | `""` | no |
| cluster\_log\_retention\_in\_days | Number of days to retain log events. Default retention - 90 days. | number | `"90"` | no |
| cluster\_name | Name of the EKS cluster. Also used as a prefix in names of related resources. | string | n/a | yes |
| cluster\_security\_group\_id | If provided, the EKS cluster will be attached to this security group. If not given, a security group will be created with necessary ingress/egress to work with the workers | string | `""` | no |
| cluster\_version | Kubernetes version to use for the EKS cluster. | string | `"1.16"` | no |
| config\_output\_path | Where to save the Kubectl config file (if `write_kubeconfig = true`). Assumed to be a directory if the value ends with a forward slash `/`. | string | `"./"` | no |
| create\_eks | Controls if EKS resources should be created (it affects almost all resources) | bool | `"true"` | no |
| eks\_oidc\_root\_ca\_thumbprint | Thumbprint of Root CA for EKS OIDC, Valid until 2037 | string | `"9e99a48a9960b14926bb7f3b02e22da2b0ab7280"` | no |
| enable\_irsa | Whether to create OpenID Connect Provider for EKS to enable IRSA | bool | `"false"` | no |
| iam\_path | If provided, all IAM roles will be created on this path. | string | `"/"` | no |
| kubeconfig\_aws\_authenticator\_additional\_args | Any additional arguments to pass to the authenticator such as the role to assume. e.g. ["-r", "MyEksRole"]. | list(string) | `[]` | no |
| kubeconfig\_aws\_authenticator\_command | Command to use to fetch AWS EKS credentials. | string | `"aws-iam-authenticator"` | no |
| kubeconfig\_aws\_authenticator\_command\_args | Default arguments passed to the authenticator command. Defaults to [token -i $cluster_name]. | list(string) | `[]` | no |
| kubeconfig\_aws\_authenticator\_env\_variables | Environment variables that should be used when executing the authenticator. e.g. { AWS_PROFILE = "eks"}. | map(string) | `{}` | no |
| kubeconfig\_name | Override the default name used for items kubeconfig. | string | `""` | no |
| manage\_aws\_auth | Whether to apply the aws-auth configmap file. | string | `"true"` | no |
| manage\_cluster\_iam\_resources | Whether to let the module manage cluster IAM resources. If set to false, cluster_iam_role_name must be specified. | bool | `"true"` | no |
| manage\_worker\_iam\_resources | Whether to let the module manage worker IAM resources. If set to false, iam_instance_profile_name must be specified for workers. | bool | `"true"` | no |
| map\_accounts | Additional AWS account numbers to add to the aws-auth configmap. See examples/basic/variables.tf for example format. | list(string) | `[]` | no |
| map\_roles | Additional IAM roles to add to the aws-auth configmap. See examples/basic/variables.tf for example format. | object | `[]` | no |
| map\_users | Additional IAM users to add to the aws-auth configmap. See examples/basic/variables.tf for example format. | object | `[]` | no |
| node\_groups | Map of map of node groups to create. See `node_groups` module's documentation for more details | any | `{}` | no |
| node\_groups\_defaults | Map of values to be applied to all node groups. See `node_groups` module's documentaton for more details | any | `{}` | no |
| permissions\_boundary | If provided, all IAM roles will be created with this permissions boundary attached. | string | `"null"` | no |
| region | The AWS region. | string | n/a | yes |
| subnets | A list of subnets to place the EKS cluster and workers within. | list(string) | n/a | yes |
| tags | A map of tags to add to all resources. | map(string) | `{}` | no |
| vpc\_id | VPC where the cluster and workers will be deployed. | string | n/a | yes |
| wait\_for\_cluster\_cmd | Custom local-exec command to execute for determining if the eks cluster is healthy. Cluster endpoint will be available as an environment variable called ENDPOINT | string | `"until curl -k -s $ENDPOINT/healthz \u003e/dev/null; do sleep 4; done"` | no |
| wait\_for\_cluster\_interpreter | Custom local-exec command line interpreter for the command to determining if the eks cluster is healthy. | list(string) | `[ "/bin/sh", "-c" ]` | no |
| worker\_additional\_security\_group\_ids | A list of additional security group ids to attach to worker instances | list(string) | `[]` | no |
| worker\_ami\_name\_filter | Name filter for AWS EKS worker AMI. If not provided, the latest official AMI for the specified 'cluster_version' is used. | string | `""` | no |
| worker\_ami\_name\_filter\_windows | Name filter for AWS EKS Windows worker AMI. If not provided, the latest official AMI for the specified 'cluster_version' is used. | string | `""` | no |
| worker\_ami\_owner\_id | The ID of the owner for the AMI to use for the AWS EKS workers. Valid values are an AWS account ID, 'self' (the current account), or an AWS owner alias (e.g. 'amazon', 'aws-marketplace', 'microsoft'). | string | `"602401143452"` | no |
| worker\_ami\_owner\_id\_windows | The ID of the owner for the AMI to use for the AWS EKS Windows workers. Valid values are an AWS account ID, 'self' (the current account), or an AWS owner alias (e.g. 'amazon', 'aws-marketplace', 'microsoft'). | string | `"801119661308"` | no |
| worker\_create\_initial\_lifecycle\_hooks | Whether to create initial lifecycle hooks provided in worker groups. | bool | `"false"` | no |
| worker\_create\_security\_group | Whether to create a security group for the workers or attach the workers to `worker_security_group_id`. | bool | `"true"` | no |
| worker\_groups | A list of maps defining worker group configurations to be defined using AWS Launch Configurations. See workers_group_defaults for valid keys. | any | `[]` | no |
| worker\_groups\_launch\_template | A list of maps defining worker group configurations to be defined using AWS Launch Templates. See workers_group_defaults for valid keys. | any | `[]` | no |
| worker\_security\_group\_id | If provided, all workers will be attached to this security group. If not given, a security group will be created with necessary ingress/egress to work with the EKS cluster. | string | `""` | no |
| worker\_sg\_ingress\_from\_port | Minimum port number from which pods will accept communication. Must be changed to a lower value if some pods in your cluster will expose a port lower than 1025 (e.g. 22, 80, or 443). | number | `"1025"` | no |
| workers\_additional\_policies | Additional policies to be added to workers | list(string) | `[]` | no |
| workers\_group\_defaults | Override default values for target groups. See workers_group_defaults_defaults in local.tf for valid keys. | any | `{}` | no |
| workers\_role\_name | User defined workers role name. | string | `""` | no |
| workers\_subnets\_type | Type of subnets to use on worker nodes: public or private | string | `"private"` | no |
| write\_kubeconfig | Whether to write a Kubectl config file containing the cluster configuration. Saved to `config_output_path`. | bool | `"true"` | no |

## Outputs

| Name | Description |
|------|-------------|
| cluster\_id |  |
| cluster\_oidc\_issuer\_url |  |

<!-- END OF PRE-COMMIT-TERRAFORM DOCS HOOK -->
