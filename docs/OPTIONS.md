# Options List for Cluster Manifests

## Global options
|  Key |  Required | Type  | Values  | Default  | Description |
|------|-----------|--------|---------|----------|----------------------------------------------|
|[name](#name)  | +         | string | any     |  n/a     | Cluster name to be used across all resources |
|[installed](#installed) | -     | bool   | `false`, `true`| `true`| Defines if cluster should be deployed or deleted, `false` would delete existing cluster |
|[cloud.provider](#cloud.provider)| + | string | `aws`, `digitalocean` |  n/a | Define cloud provider |
|[cloud.region](#cloud.region)| + | string| region name | n/a | Define cloud provider region to create cluster |
|[cloud.domain](#cloud.domain)| - | string| `cluster.dev`, `FQDN` | cluster.dev | To expose cluster resources the DNS zone is required. If not set the installer would create a zone `cluster-name-organization.cluster.dev` and point it to cloud service NS'es. So you can use it. Alternate you can set your zone which already exist in target cloud.|


## AWS Specific Keys

