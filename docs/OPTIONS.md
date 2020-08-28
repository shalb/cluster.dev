# Options List for Cluster Manifests

The section contains a list of options that are set in cluster manifests (yaml files) to configure the clusters.

## Manifest Example
*For more examples please see the [/.cluster.dev](https://github.com/shalb/cluster.dev/tree/master/.cluster.dev) directory.*

```yaml
# .cluster.dev/staging.yaml
cluster:
  installed: true
  name: staging
  cloud:
    provider: aws
    region: eu-central-1
    vpc: default
    domain: cluster.dev
    provisioner:
      type: minikube
      instanceType: m5.large
  addons:
    nginx-ingress: true
    cert-manager: true
  apps:
    - /kubernetes/apps/samples
```

## Global Options

|  Key |  Required | Type  | Values  | Default  | Description |
|------|-----------|--------|---------|----------|----------------------------------------------|
|[name](#name)  | + | string | any     | | Cluster name to be used across all resources. |
|[installed](#installed) | - | bool| `false`, `true`| `true`| Defines if the cluster should be deployed or deleted, `false` would delete the existing cluster. |
|[cloud.provider](#cloud.provider)| + | string | `aws`, `digitalocean` | | Defines a cloud provider. |
|[cloud.region](#cloud.region)| + | string |  ex: `us-east-1`, `ams3` | | Defines a cloud provider region, in which to create the cluster. |
|[cloud.availability_zones](#cloud.availability_zones)| - | string | ex:  `us-east-1b`, `us-east-1c`| `cloud.region'a'`,`cloud.region'b'`| Define networks and nodes location inside a single region. Minimum two zones should be defined. Cluster nodes could be spread across different datacenters. Multiple zones provide high availability, however, can lead to cost increase. |
|[cloud.domain](#cloud.domain)| - | string| FQDN ex: `cluster.dev`, `example.org` | `cluster.dev` | To expose cluster resources, the DNS zone is required. If set to `cluster.dev`, the installer would create a zone `cluster-name-organization.cluster.dev` and point it to your cloud service NS'es. Alternate, you can set your own zone, which already exists in the target cloud. |
|[cloud.vpc](#cloud.vpc)| - |string|`default`, `create`, `vpc_id`| `default`| Virtual Private Cloud. `default` - use default one, `create` - installer creates a new VPC, `vpc_id` - define an already existent (in AWS tag the networks manually with the "cluster.dev/subnet_type" = "private/public" tags to make them usable). |
|[cloud.vpc_cidr](#cloud.vpc)| - |string| ex:`10.2.0.0/16`, `192.168.0.0/20`| `10.8.0.0/18`| The CIDR block for the VPC. Cluster pods will use IPs from that pool. If you need peering between VPCs, their CIDRs should be unique. |


## Cluster Provisioners

### Amazon AWS Provisioners

Required environment variables to be passed to the container:

 - `AWS_ACCESS_KEY_ID` - the access key ID required for user programmatic access, see the [AWS documentation](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_users_create.html).

- `AWS_SECRET_ACCESS_KEY` - the secret access key required for user programmatic access.

|  Key |  Required | Type  | Values  | Default  | Description |
|------|-----------|--------|---------|----------|----------------------------------------------|
| [provisioner.type](#provisioner.type) | + | string | `minikube`, `eks` | | Provisioner to deploy the cluster with. |

#### AWS Minikube

*example yaml file: [.cluster.dev/aws-minikube.yaml](https://github.com/shalb/cluster.dev/blob/master/.cluster.dev/aws-minikube.yaml)*

|  Key |  Required | Type  | Values  | Default  | Description |
|------|-----------|--------|---------|----------|----------------------------------------------|
| [provisioner.instanceType](#provisioner.instanceType) | + | string | ex:`m5.xlarge` | `m5.large` | Single node Kubernetes cluster AWS EC2 instance type. |

#### AWS EKS

*example yaml file: [.cluster.dev/aws-eks.yaml](https://github.com/shalb/cluster.dev/blob/master/.cluster.dev/aws-eks.yaml)*

|  Key |  Required | Type  | Values &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;| Default  |Description &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;  |
|------|-----------|--------|---------|----------|----------------------------------------------|
| [provisioner.version](#provisioner.version) | + | string | ex:`1.15`, `1.16` | `1.16` | Kubernetes version. |
| [provisioner.additional_security_group_ids](#provisioner.node_group.additional_security_group_ids) | - | list | ex:`sg-233ba1, sg-2221bb` |  |  A list of additional security group IDs to include in worker launch config. |
| [provisioner.node_group.name](#provisioner.node_group.name) | + | string | ex:`spot-group`, `on-demand-group` | `node-group` | Name for Kubernetes group of worker nodes. |
| [provisioner.node_group.type](#provisioner.node_group.type) | - | string | ex:`spot`, `on-demand` | `on-demand` | Type for Kubernetes group of worker nodes. |
| [provisioner.node_group.instance_type](#provisioner.node_group.instance_type) | - | string | ex:`t3.medium`, `c5.xlarge` | `m5.large` | Size of a Kubernetes worker node. |
| [provisioner.node_group.asg_desired_capacity](#provisioner.node_group.asg_desired_capacity) | - | integer | `1`..<=`asg_max_size` | `1` | Desired worker capacity in the autoscaling group. |
| [provisioner.node_group.asg_max_size](#provisioner.node_group.asg_max_size) | - | integer | `1`..`cloud limit` | `3` | Maximum worker capacity in the autoscaling group. |
| [provisioner.node_group.asg_min_size](#provisioner.node_group.asg_min_size) | - | integer | `1`..<=`asg_max_size` | `1` | Minimum worker capacity in the autoscaling group. |
| [provisioner.node_group.root_volume_size](#provisioner.node_group.root_volume_size) | - | integer | `20`..`cloud limit` | `40` | Root volume size in GB in worker instances. |
| [provisioner.node_group.kubelet_extra_args](#provisioner.node_group.kubelet_extra_args) | - | string | `--node-labels=kubernetes.io/lifecycle=spot` |  |  This string is passed directly to kubelet, if set. Useful for adding labels or taints. |
| [provisioner.node_group.override_instance_types](#provisioner.node_group.override_instance_types) | - | list | ex:`m5.large, m5a.large, c5.xlarge` |  | A list of override instance types for mixed instances policy. |
| [provisioner.node_group.spot_allocation_strategy](#provisioner.node_group.spot_allocation_strategy) | - | string | `lowest-price`, `capacity-optimized` | `lowest-price` |  If 'lowest-price', the Auto Scaling group launches instances using the Spot pools with the lowest price, and evenly allocates your instances across the number of Spot pools.  If 'capacity-optimized', the Auto Scaling group launches instances using Spot pools that are optimally chosen based on the available Spot capacity. |
| [provisioner.node_group.spot_instance_pools](#provisioner.node_group.spot_instance_pools) | - | integer | `1`..`cloud limit` | `10` | Number of Spot pools per availability zone to allocate capacity. EC2 Auto Scaling selects the cheapest Spot pools and evenly allocates Spot capacity across the number of Spot pools that you specify. |
| [provisioner.node_group.spot_max_price](#provisioner.node_group.spot_max_price) | - | float | "" | "" | Maximum price per unit hour that the user is willing to pay for the Spot instances. Default is the on-demand price. |
| [provisioner.node_group.on_demand_base_capacity](#provisioner.node_group.on_demand_base_capacity) | - | integer | `0`..`100` | `0` | Absolute minimum amount of desired capacity that must be fulfilled by on-demand instances. |
| [provisioner.node_group.on_demand_percentage_above_base_capacity](#provisioner.node_group.on_demand_percentage_above_base_capacity) | - | integer | `0`..`100` | `0` | Percentage split between on-demand and Spot instances above the base on-demand capacity. |


### DigitalOcean Provisioners

The DigitalOcean (DO) provider is used to interact with the resources supported by DigitalOcean. The following environment variables should be set:

- `DIGITALOCEAN_TOKEN` - the DO API token, see the [DO documentation](https://www.digitalocean.com/docs/apis-clis/api/create-personal-access-token/).
- `SPACES_ACCESS_KEY_ID` - the access key ID used for Spaces API operations, see the [DO documentation](https://www.digitalocean.com/community/tutorials/how-to-create-a-digitalocean-space-and-api-key).
- `SPACES_SECRET_ACCESS_KEY` - the secret access key used for Spaces API operations.


|  Key |  Required | Type  | Values  | Default  | Description |
|------|-----------|--------|---------|----------|----------------------------------------------|
| [provisioner.type](#provisioner.type) | + | string | `managed-kubernetes` | | Provisioner to deploy cluster with.|
| [cloud.project](#cloud.project) | - | string | ex: `staging` | `default` | DigitalOcean Project name.|


#### DigitalOcean Managed Kubernetes

*example yaml file: [.cluster.dev/digitalocean-k8s.yaml](https://github.com/shalb/cluster.dev/blob/master/.cluster.dev/digitalocean-k8s.yaml)*

|  Key |  Required | Type  | Values&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;  | Default  &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;| Description |
|------|-----------|--------|---------|----------|----------------------------------------------|
| [version](#version) | - | string | ex: `1.16` | | DigitalOcean managed Kubernetes [version](https://www.digitalocean.com/docs/kubernetes/changelog/). |
| [nodeCount](#nodeCount) | + | integer | `1-512`  | `1` | Number of Droplet instances in the cluster. |
| [nodeSize](#nodeSize) | + | string | ex: `s-4vcpu-8gb`  | `s-1vcpu-2gb` | The slug identifier for the type of Droplets used as workers in the node pool. |
| [autoScale](#autoScale) | - | boolean | `true`, `false`  | `false` | A boolean indicating whether auto-scaling is enabled.|
| [minNodes](#minNodes) | - | boolean |`1-512` | `1` | If `autoScale` is enabled, defines a minimum number of Droplet instances in the cluster. |
| [maxNodes](#maxNodes) | - | boolean |`1-512` | `1` | If `autoScale` is enabled, defines a maximum number of Droplet instances in the cluster. |

## Cluster Addons
|  Key |  Required | Type  | Values  | Default  | Description |
|------|-----------|--------|---------|----------|----------------------------------------------|
| [nginx-ingress](#nginx-ingress) | - | boolean | `true`,`false` | `true` | Deploy [nginx-ingress](https://github.com/kubernetes/ingress-nginx). |
| [cert-manager](#cert-manager) | - | boolean | `true`,`false` | `true` | Deploy [cert-manager](https://cert-manager.io/). |
| [external-dns](#external-dns) | - | boolean | `true`,`false` | `true` | Deploy [external-dns](https://github.com/kubernetes-sigs/external-dns/). |
| [argo-cd](#argo-cd) | - | boolean | `true`,`false` | `true` | Deploy [argo-cd](https://argoproj.github.io/argo-cd/). |
| [olm](#olm) | - | boolean | `true`,`false` | `true` | Deploy [Operator Lifecycle Manager](https://github.com/operator-framework/operator-lifecycle-manager). |
| [keycloak](#keycloak) | - | boolean | `true`,`false` | `true` | Deploy [Keycloak Operator](https://github.com/keycloak/keycloak-operator). |
