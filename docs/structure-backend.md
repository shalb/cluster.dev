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

Stores the cluster state in AWS S3 bucket. The Cluster.dev S3 backend supports some options of [Terraform S3](https://developer.hashicorp.com/terraform/language/settings/backends/s3) backend. The list of supported options is referred below. 

```yaml
name: aws-backend
kind: backend
provider: s3
spec:
  bucket: cdev-states
  region: {{ .project.variables.region }}
```

#### Options

* `bucket` - *required*. The name of the S3 bucket. 

* `region` - *required*. AWS Region of the S3 Bucket and DynamoDB Table (if used). This can also be sourced from the `AWS_DEFAULT_REGION` and `AWS_REGION` environment variables.

* `access_key` - *optional*. AWS access key. If configured, must also configure `secret_key`. This can also be sourced from the `AWS_ACCESS_KEY_ID` environment variable, AWS shared credentials file (e.g. ~/.aws/credentials), or AWS shared configuration file (e.g. ~/.aws/config).

* `secret_key` - *optional*. AWS access key. If configured, must also configure `access_key`. This can also be sourced from the `AWS_SECRET_ACCESS_KEY` environment variable, AWS shared credentials file (e.g. ~/.aws/credentials), or AWS shared configuration file (e.g. ~/.aws/config).

* `profile` - *optional*. Name of AWS profile in AWS shared credentials file (e.g. ~/.aws/credentials) or AWS shared configuration file (e.g. ~/.aws/config) to use for credentials and/or configuration. This can also be sourced from the `AWS_PROFILE` environment variable.

* `token` - *optional*. Multi-Factor Authentication (MFA) token. This can also be sourced from the `AWS_SESSION_TOKEN` environment variable.

* `endpoint` - *optional*. Custom endpoint URL for the AWS S3 API.

* `skip_metadata_api_check` - *optional*. Skip usage of EC2 Metadata API.

* `skip_credentials_validation` - *optional*. Skip credentials validation via the STS API. Useful for testing and for AWS API implementations that do not have STS available.

* `max_retries` - *optional*. The maximum number of times an AWS API request is retried on retryable failure. Defaults to 5.

* `shared_credentials_file` - *optional*, *deprecated*, use `shared_credentials_files` instead. Path to the AWS shared credentials file. Defaults to ~/.aws/credentials.

* `skip_region_validation` - *optional*. Skip validation of provided region name.

* `sts_endpoint` - *optional*, *deprecated*. Custom endpoint URL for the AWS Security Token Service (STS) API. Use `endpoints.sts` instead.

* `iam_endpoint` - *optional*, *deprecated*. Custom endpoint URL for the AWS Identity and Access Management (IAM) API. Use `endpoints.iam` instead.

* `force_path_style` - *optional*, *deprecated*. Enable path-style S3 URLs (https://<HOST>/<BUCKET> instead of https://<BUCKET>.<HOST>).

* `assume_role_policy` - *optional*. IAM Policy JSON describing further restricting permissions for the IAM Role being assumed. Use `assume_role.policy` instead.

* `assume_role_policy_arns` - *optional*. Set of Amazon Resource Names (ARNs) of IAM Policies describing further restricting permissions for the IAM Role being assumed. Use `assume_role.policy_arns` instead.

* `assume_role_tags` - *optional*. Map of assume role session tags. Use `assume_role.tags` instead.

* `assume_role_transitive_tag_keys` - *optional*. Set of assume role session tag keys to pass to any subsequent sessions. Use `assume_role.transitive_tag_keys` instead.

* `external_id` - *optional*. External identifier to use when assuming the role. Use `assume_role.external_id` instead.

* `role_arn` - *optional*. Amazon Resource Name (ARN) of the IAM Role to assume. Use `assume_role.role_arn` instead.

* `session_name` - *optional*. Session name to use when assuming the role. Use `assume_role.session_name` instead.

### `azurerm`

Stores the cluster state in Microsoft Azure cloud. The `azurerm` backend supports the options of [Terraform azurerm](https://www.terraform.io/language/settings/backends/azurerm) backend.

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

Stores the cluster state in Google Cloud service. The `gcs` backend supports some options of [Terraform gcs](https://www.terraform.io/language/settings/backends/gcs) backend. The list of supported options is referred below. 

```yaml
name: gcs-b
kind: backend
provider: gcs
spec:
  bucket: cdev-states
  prefix: pref
```

#### Options

* `bucket` - *required*. The name of the GCS bucket. This name must be globally unique. For more information, see [Bucket Naming Guidelines](https://cloud.google.com/storage/docs/buckets#naming).

* `credentials` / `GOOGLE_BACKEND_CREDENTIALS` / `GOOGLE_CREDENTIALS` - *optional*. Local path to Google Cloud Platform account credentials in JSON format. If unset, the path uses [Google Application Default Credentials](https://cloud.google.com/docs/authentication/application-default-credentials). The provided credentials must have the Storage Object Admin role on the bucket. **Warning**: if using the Google Cloud Platform provider as well, it will also pick up the `GOOGLE_CREDENTIALS` environment variable.

* `impersonate_service_account` / GOOGLE_BACKEND_IMPERSONATE_SERVICE_ACCOUNT / GOOGLE_IMPERSONATE_SERVICE_ACCOUNT - *optional*. The service account to impersonate for accessing the State Bucket. You must have `roles/iam.serviceAccountTokenCreator` role on that account for the impersonation to succeed. If you are using a delegation chain, you can specify that using the `impersonate_service_account_delegates` field.

* `impersonate_service_account_delegates` - *optional*. The delegation chain for an impersonating a service account as described [here](https://cloud.google.com/iam/docs/create-short-lived-credentials-direct#sa-credentials-delegated).

* `access_token` - *optional*. A temporary [OAuth 2.0 access token] obtained from the Google Authorization server, i.e. the `Authorization: Bearer` token used to authenticate HTTP requests to GCP APIs. This is an alternative to credentials. If both are specified, `access_token` will be used over the credentials field.

* `prefix` - *optional*. GCS prefix inside the bucket. Named states for workspaces are stored in an object called `<prefix>/<name>.tfstate`.

* `encryption_key` / GOOGLE_ENCRYPTION_KEY - *optional*. A 32 byte base64 encoded 'customer-supplied encryption key' used when reading and writing state files in the bucket. For more information see [Customer-supplied Encryption Keys](https://cloud.google.com/storage/docs/encryption/customer-supplied-keys).

* `storage_custom_endpoint` / GOOGLE_BACKEND_STORAGE_CUSTOM_ENDPOINT / GOOGLE_STORAGE_CUSTOM_ENDPOINT - *optional*. A URL containing three parts: the protocol, the DNS name pointing to a Private Service Connect endpoint, and the path for the Cloud Storage API (`/storage/v1/b`, see [here](https://cloud.google.com/storage/docs/json_api/v1/buckets/get#http-request)). You can either use [a DNS name automatically made by the Service Directory](https://cloud.google.com/vpc/docs/configure-private-service-connect-apis#configure-p-dns) or a [custom DNS name](https://cloud.google.com/vpc/docs/configure-private-service-connect-apis#configure-dns-default) made by you. For example, if you create an endpoint called `xyz` and want to use the automatically-created DNS name, you should set the field value as `https://storage-xyz.p.googleapis.com/storage/v1/b`. For help creating a Private Service Connect endpoint using Terraform, see [this guide](https://cloud.google.com/vpc/docs/configure-private-service-connect-apis#terraform_1).

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

