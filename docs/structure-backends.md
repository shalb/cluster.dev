# Backend

Backend is an object that describes backend storage for Terraform and Cluster.dev states.

File: searching in `./*.yaml`. *Required at least one*.
In the backends' configuration you can use any options of the appropriate Terraform backend. They will be converted as is.
Currently 4 types of backends are supported:

* `s3` AWS S3 backend:

```yaml
name: aws-backend
kind: backend
provider: s3
spec:
  bucket: cdev-states
  region: {{ .project.variables.region }}
```

* `do` DigitalOcean spaces backend:

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

* `azurerm` Microsoft azurerm:

```yaml
name: gcs-b
kind: backend
provider: azurerm
spec:
  resource_group_name: "StorageAccount-ResourceGroup"
  storage_account_name: "example"
  container_name: "cdev-states"
```

* `gcs` Google Cloud backend:

```yaml
name: do-backend
kind: backend
provider: gcs
spec:
  bucket: cdev-states
  prefix: pref
```
