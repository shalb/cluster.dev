# Backends

File: searching in `./*.yaml`. *Optional*.

Backend is an object that describes backend storage for Terraform and Cluster.dev [states](https://docs.cluster.dev/cluster-state/). A backend could be [local](#local-backend) or [remote](#remote-backend), depending on where it stores a state.  

 You can use any options of [Terraform backends](https://www.terraform.io/language/settings/backends) in the remote backend configuration. The options will be mapped to a generated Terraform backend and converted as is.

## Local backend

Local backend stores the cluster state on a local file system in the `.cluster.dev/states/cdev-state.json` file. Cluster.dev will use the local backend by default unless the remote backend is specified in the `project.yaml`. 

Example configuration:

```yaml
name: my-fs
kind: backend
provider: local
spec:
  path: /home/cluster.dev/states/
```

A path should be absolute or relative to the directory where `cdev` is running. An absolute path must begin with `/`, and a relative with `./` or `../`. 

## Remote backend

Remote backend uses remote cloud services to store the cluster state, making it accessible for team work.

### `s3`

Stores the cluster state in AWS S3 bucket. The S3 backend supports all options of [Terraform S3](https://developer.hashicorp.com/terraform/language/settings/backends/s3) backend.

```yaml
name: aws-backend
kind: backend
provider: s3
spec:
  bucket: cdev-states
  region: {{ .project.variables.region }}
```

### `azurerm`

Stores the cluster state in Microsoft Azure cloud. The `azurerm` backend supports all options of [Terraform azurerm](https://www.terraform.io/language/settings/backends/azurerm) backend.

```yaml
name: azurerm-b
kind: backend
provider: azurerm
spec:
  resource_group_name: "StorageAccount-ResourceGroup"
  storage_account_name: "example"
  container_name: "cdev-states"
```

### `gcs`

Stores the cluster state in Google Cloud service. The `gcs` backend supports all options of [Terraform gcs](https://www.terraform.io/language/settings/backends/gcs) backend. 

```yaml
name: gcs-b
kind: backend
provider: gcs
spec:
  bucket: cdev-states
  prefix: pref
```

### Digital Ocean Spaces and MinIO

To use DO spaces or MinIO object storage as a backend, use `s3` backend provider with additional options. See details: 

- [DO Spaces](https://anichakraborty.medium.com/terraform-remote-state-backup-with-digital-ocean-spaces-697e35128a6a)
- [MinIO](https://ruben-rodriguez.github.io/posts/minio-s3-terraform-backend/)

DO Spaces example:

```yaml
name: do-backend
kind: Backend
provider: s3
spec:
  bucket: cdev-state
  region: main
  access_key: "<SPACES_SECRET_KEY>" # Optional, it's better to use environment variable 'export SPACES_SECRET_KEY="key"'
  secret_key: "<SPACES_ACCESS_TOKEN>" # Optional, it's better to use environment variable 'export SPACES_ACCESS_TOKEN="token"'
  endpoint: "https://sgp1.digitaloceanspaces.com"
  skip_credentials_validation: true
  skip_region_validation: true
  skip_metadata_api_check: true
```

MinIO example:

```yaml
name: minio-backend
kind: Backend
provider: s3
spec:
  bucket: cdev-state
  region: main
  access_key: "minioadmin"
  secret_key: "minioadmin"
  endpoint: http://127.0.0.1:9000
  skip_credentials_validation: true
  skip_region_validation: true
  skip_metadata_api_check: true
  force_path_style: true
```

