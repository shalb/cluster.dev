# Getting Started with Cluster.dev on AWS

This guide will walk you through the steps to deploy your first project with Cluster.dev on AWS.

## Prerequisites

Ensure the following are installed and set up:

- **Terraform**: Version 1.4 or above. [Install Terraform.](https://developer.hashicorp.com/terraform/downloads)
  ```bash
  terraform --version
  ```

- **AWS CLI**:
  ```bash
  curl "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv2.zip"
  unzip awscliv2.zip
  sudo ./aws/install
  aws --version
  ```

- **Cluster.dev client**:
  ```bash
  curl -fsSL https://raw.githubusercontent.com/shalb/cluster.dev/master/scripts/get_cdev.sh | sh
  cdev --version
  ```

## Authentication

Choose one of the two methods below:

1. **Shared Credentials File** (recommended):

    - Populate `~/.aws/credentials`:

        ```bash
        [cluster-dev]
        aws_access_key_id = YOUR_AWS_ACCESS_KEY
        aws_secret_access_key = YOUR_AWS_SECRET_KEY
        ```

    - Configure `~/.aws/config`:

        ```bash
        [profile cluster-dev]
        region = eu-central-1
        ```

    - Set the AWS profile:

        ```bash
        export AWS_PROFILE=cluster-dev
        ```

2. **Environment Variables**:
   ```bash
   export AWS_ACCESS_KEY_ID="YOUR_AWS_ACCESS_KEY"
   export AWS_SECRET_ACCESS_KEY="YOUR_AWS_SECRET_KEY"
   export AWS_DEFAULT_REGION="eu-central-1"
   ```
## Creating an S3 Bucket for Storing State

```bash
aws s3 mb s3://cdev-states
```
## Setting Up Your Project

### Project Configuration (`project.yaml`)

*   Defines the overarching project settings. All subsequent stack configurations will inherit and can override these settings.
*   It points to aws-backend as the backend, meaning that the Cluster.dev state for resources defined in this project will be stored in the S3 bucket specified in `backend.yaml`.
*   Project-level variables are defined here and can be referenced in other configurations.

```bash
cat <<EOF > project.yaml
name: dev
kind: Project
backend: aws-backend
variables:
  organization: cluster.dev
  region: eu-central-1
  state_bucket_name: cdev-states
EOF
```

### Backend Configuration (`backend.yaml`)

This specifies where Cluster.dev will store its own state and the Terraform states for any infrastructure it provisions or manages. Given the backend type as S3, it's clear that AWS is the chosen cloud provider.

```bash
cat <<EOF > backend.yaml
name: aws-backend
kind: Backend
provider: s3
spec:
  bucket: {{ .project.variables.state_bucket_name }}
  region: {{ .project.variables.region }}
EOF
```

### Stack Configuration (`stack.yaml`)

*   This represents a distinct set of infrastructure resources to be provisioned.
*   It references a local template (in this case, the previously provided stack template) to know what resources to create.
*   Variables specified in this file will be passed to the Terraform modules called in the template.
*   The content variable here is especially useful; it dynamically populates the content of an S3 bucket object (a webpage in this case) using the local `index.html` file.

```bash
cat <<EOF > stack.yaml
name: s3-website
template: ./template/
kind: Stack
backend: aws-backend
variables:
  bucket_name: "tmpl-dev-test"
  region: {{ .project.variables.region }}
  content: |
    {{- readFile "./files/index.html" | nindent 4 }}
EOF
```

### Stack Template (`template.yaml`)

The `StackTemplate` serves as a pivotal object within Cluster.dev. It lays out the actual infrastructure components you intend to provision using Terraform modules and resources. Essentially, it determines how your cloud resources should be laid out and interconnected.

```bash
mkdir template
cat <<EOF > template/template.yaml
_p: &provider_aws
- aws:
    region: {{ .variables.region }}

name: s3-website
kind: StackTemplate
units:
  -
    name: bucket
    type: tfmodule
    providers: *provider_aws
    source: terraform-aws-modules/s3-bucket/aws
    inputs:
      bucket: {{ .variables.bucket_name }}
      force_destroy: true
      acl: "public-read"
      control_object_ownership: true
      object_ownership: "BucketOwnerPreferred"
      attach_public_policy: true
      block_public_acls: false
      block_public_policy: false
      ignore_public_acls: false
      restrict_public_buckets: false
      website:
        index_document: "index.html"
        error_document: "error.html"
  -
    name: web-page-object
    type: tfmodule
    providers: *provider_aws
    source: "terraform-aws-modules/s3-bucket/aws//modules/object"
    version: "3.15.1"
    inputs:
      bucket: {{ remoteState "this.bucket.s3_bucket_id" }}
      key: "index.html"
      acl: "public-read"
      content_type: "text/html"
      content: |
        {{- .variables.content | nindent 8 }}

  -
    name: outputs
    type: printer
    depends_on: this.web-page-object
    outputs:
      websiteUrl: http://{{ .variables.bucket_name }}.s3-website.{{ .variables.region }}.amazonaws.com/
EOF
```

<details>
  <summary>Click to expand explanation of the Stack Template</summary>

 <h4>1. Provider Definition (_p)</h4> <br>

This section employs a YAML anchor, pre-setting the cloud provider and region for the resources in the stack. For this example, AWS is the designated provider, and the region is dynamically passed from the variables:

```yaml
_p: &provider_aws
- aws:
    region: {{ .variables.region }}
```

<h4>2. Units</h4> <br>

The units section is where the real action is. Each unit is a self-contained "piece" of infrastructure, typically associated with a particular Terraform module or a direct cloud resource. <br>

&nbsp;  

<h5>Bucket Unit</h5> <br>

This unit is utilizing the `terraform-aws-modules/s3-bucket/aws` module to provision an S3 bucket. Inputs for the module, such as the bucket name, are populated using variables passed into the Stack.

```yaml
name: bucket
type: tfmodule
providers: *provider_aws
source: terraform-aws-modules/s3-bucket/aws
inputs:
  bucket: {{ .variables.bucket_name }}
  ...
```

<h5>Web-page Object Unit</h5> <br>

After the bucket is created, this unit takes on the responsibility of creating a web-page object inside it. This is done using a sub-module from the S3 bucket module specifically designed for object creation. A notable feature is the remoteState function, which dynamically pulls the ID of the S3 bucket created by the previous unit:

```yaml
name: web-page-object
type: tfmodule
providers: *provider_aws
source: "terraform-aws-modules/s3-bucket/aws//modules/object"
inputs:
  bucket: {{ remoteState "this.bucket.s3_bucket_id" }}
  ...
```

<h5>Outputs Unit</h5> <br>

Lastly, this unit is designed to provide outputs, allowing users to view certain results of the Stack execution. For this template, it provides the website URL of the hosted S3 website.

```yaml
name: outputs
type: printer
depends_on: this.web-page-object
outputs:
  websiteUrl: http://{{ .variables.bucket_name }}.s3-website.{{ .variables.region }}.amazonaws.com/
```

<h4>3. Variables and Data Flow</h4> <br>

The Stack Template is adept at harnessing variables, not just from the Stack (e.g., `stack.yaml`), but also from other resources via the remoteState function. This facilitates a seamless flow of data between resources and units, enabling dynamic infrastructure creation based on real-time cloud resource states and user-defined variables.
</details>

### Sample Website File (`files/index.html`)

```bash
mkdir files
cat <<EOF > files/index.html
<html xmlns="http://www.w3.org/1999/xhtml">
<head>
    <title>Cdev Demo Website Home Page</title>
</head>
<body>
  <h1>Welcome to my website</h1>
  <p>Now hosted on Amazon S3!</p>
  <h2>See you!</h2>
</body>
</html>
EOF
```

## Deploying with Cluster.dev

- Plan the deployment:

    ```bash
    cdev plan
    ```

- Apply the changes:

    ```bash
    cdev apply
    ```

### Example Screen Cast

![asciicast](images/611594.cast)
## Clean up

To remove the cluster with created resources run the command:

```bash
cdev destroy
```
