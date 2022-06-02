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
  path: /home/cluster.dev/states/cdev-state.json
```

## Remote backend

Remote backend uses remote cloud services to store the cluster state, making it accessible for team work. 

Currently you can use only S3 bucket as a remote backend. In the future we plan to add other remote backend options that are listed below. 

### `s3` 

Stores the cluster state in AWS S3 bucket. 

```yaml
name: aws-backend
kind: backend
provider: s3
spec:
  bucket: cdev-states
  region: {{ .project.variables.region }}
```

### `azurerm` 

Stores the cluster state in Microsoft Azure cloud. You can also use any options of [Terraform azurerm](https://www.terraform.io/language/settings/backends/azurerm) backend.

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

Stores the cluster state in Google Cloud service. You can also use any options of [Terraform gcs](https://www.terraform.io/language/settings/backends/gcs) backend. 

```yaml
name: gcs-b
kind: backend
provider: gcs
spec:
  bucket: cdev-states
  prefix: pref
```

### `do` 

Stores the cluster state in DigitalOcean spaces. 

```yaml
name: do-backend
kind: backend
provider: do
spec:
  bucket: cdev-states
  region: {{ .project.variables.region }}
  access_key: {{ env "SPACES_ACCESS_KEY_ID" }}
  secret_key: {{ env "SPACES_SECRET_ACCESS_KEY" }}
```