# **Getting Started with Cluster.dev on Google Cloud**

This guide will walk you through the steps to deploy your first project with Cluster.dev on Google Cloud.

```text
                          +---------------------------------+
                          | Project.yaml                    |
                          |  - project_name                 |
                          |  - google_project_id            |
                          |  - google_cloud_region          |
                          |  - google_cloud_bucket_location |
                          +------------+--------------------+
                                       |
                                       |
                          +------------v------------+
                          | Stack.yaml              |
                          |  - web_page_content     |
                          +------------+------------+
                                       |
                                       |
+--------------------------------------v-----------------------------------------------------------------+
| StackTemplate: gcs-static-website                                                                      |
|                                                                                                        |
|  +---------------------+     +---------------------+     +-----------------+    +-----------------+    |
|  | cloud-storage       |     | cloud-bucket-object |     | cloud-url-map   |    | cloud-lb        |    |
|  | type: tfmodule      |     | type: tfmodule      |     | type: tfmodule  |    | type: tfmodule  |    |
|  | inputs:             |     | inputs:             |     | inputs:         |    | inputs:         |    |
|  |  names              |     |   bucket_name       |     |  name           |    |  name           |    |
|  |  randomize_suffix   |     |   object_name       |     |  bucket_name    |    |  project        |    |
|  |  project_id         |     |   object_content    |     +--------^--------+    |  url_map        |    |
|  |  location           |     +----------^----------+      |                     +--------^--------+    |
|  +---------------------+       |                          |                       |                    |
|        |                       | cloud-storage            | cloud-storage         | cloud-url-map      |
|        |                       | bucket name              | bucket name           | url_map            |
|        |                       | via remoteState          | via remoteState       | via remoteState    |
+--------|-----------------------|--------------------------|--------------------------------------------+
         |                       |                          |                       |
         v                       v                          v                       v
  Storage Bucket             Storage Bucket Object     Url Map & Bucket Backend   Load Balancer
                                 |
                                 v
                       Printer: Static WebsiteUrl
```

## Prerequisites

Ensure the following are installed and set up:

- **Terraform**: Version 1.4 or above. [Install Terraform.](https://developer.hashicorp.com/terraform/downloads)

  ```bash
  terraform --version
  ```

- **Google Cloud CLI**: [Install Google Cloud CLI.](https://cloud.google.com/sdk/docs/install)

  ```bash
  gcloud --version
  ```

- **Cluster.dev client**:

  ```bash
  curl -fsSL https://raw.githubusercontent.com/shalb/cluster.dev/master/scripts/get_cdev.sh | sh
  cdev --version
  ```

## Authentication

Before using the Google Cloud  CLI, you'll need to authenticate:

  ```bash
  gcloud auth login
  ```

Authorize cdev/terraform to interact with GCP via SD

  ```bash
  gcloud auth application-default login
  ```

## Creating an Storage Bucket for Storing State


  ```bash
  gsutil mb gs://cdevstates
  ```

## Setting Up Your Project

### Project Configuration (`project.yaml`)

- Defines the overarching project settings. All subsequent stack configurations will inherit and can override these settings.
- It points to aws-backend as the backend, meaning that the Cluster.dev state for resources defined in this project will be stored in the Google Storage bucket specified in `backend.yaml`.
- Project-level variables are defined here and can be referenced in other configurations.

```bash
cat <<EOF > project.yaml
name: dev
kind: Project
backend: gcs-backend
variables:
  project_name: dev-test
  google_project_id: cdev-demo
  google_cloud_region: us-west1
  google_cloud_bucket_location: EU
  google_bucket_name: cdevstates
  google_bucket_prefix: dev
EOF
```

### Backend Configuration (`backend.yaml`)

This specifies where Cluster.dev will store its own state and the Terraform states for any infrastructure it provisions or manages. Given the backend type as GCS.

```bash
name: gcs-backend
kind: Backend
provider: gcs
spec:
  project: {{ .project.variables.google_project_id }}
  bucket: {{ .project.variables.google_bucket_name }}
  prefix: {{ .project.variables.google_bucket_prefix }}
EOF
```

### Stack Configuration (`stack.yaml`)

*   This represents a distinct set of infrastructure resources to be provisioned.
*   It references a local template (in this case, the previously provided stack template) to know what resources to create.
*   Variables specified in this file will be passed to the Terraform modules called in the template.
*   The content variable here is especially useful; it dynamically populates the content of an Google Storage bucket object (a webpage in this case) using the local `index.html` file.

```bash
cat <<EOF > stack.yaml
name: cloud-storage
template: ./template/
kind: Stack
backend: gcs-backend
variables:
  project_name: {{ .project.variables.project_name }}
  google_cloud_region: {{ .project.variables.google_cloud_region }}
  google_cloud_bucket_location: {{ .project.variables.google_cloud_bucket_location }}
  google_project_id: {{ .project.variables.google_project_id }}
  web_page_content: |
    {{- readFile "./files/index.html" | nindent 4 }}
EOF
```

### Stack Template (`template.yaml`)

The `StackTemplate` serves as a pivotal object within Cluster.dev. It lays out the actual infrastructure components you intend to provision using Terraform modules and resources. Essentially, it determines how your cloud resources should be laid out and interconnected.

```bash
mkdir template
cat <<EOF > template/template.yaml
_p: &provider_gcp
- google:
    project: {{ .variables.google_project_id }}
    region: {{ .variables.google_cloud_region }}

name: gcs-static-website
kind: StackTemplate
units:
  -
    name: cloud-storage
    type: tfmodule
    providers: *provider_gcp
    source: "github.com/terraform-google-modules/terraform-google-cloud-storage.git?ref=v4.0.1"
    inputs:
      names:
        - {{ .variables.project_name }}
      randomize_suffix: true
      project_id: {{ .variables.google_project_id }}
      location: {{ .variables.google_cloud_bucket_location }}
      set_viewer_roles: true
      viewers:
        - allUsers
      website:
        main_page_suffix: "index.html"
        not_found_page: "index.html"
  -
    name: cloud-bucket-object
    type: tfmodule
    providers: *provider_gcp
    depends_on: this.cloud-storage
    source: "bootlabstech/cloud-storage-bucket-object/google"
    version: "1.0.1"
    inputs:
      bucket_name: {{ remoteState "this.cloud-storage.name" }}
      object_name: "index.html"
      object_content: |
        {{- .variables.web_page_content | nindent 8 }}
  -
    name: cloud-url-map
    type: tfmodule
    providers: *provider_gcp
    depends_on: this.cloud-storage
    source: "github.com/shalb/terraform-gcs-bucket-backend.git?ref=0.0.1"
    inputs:
      name: {{ .variables.project_name }}
      bucket_name: {{ remoteState "this.cloud-storage.name" }}
  -
    name: cloud-lb
    type: tfmodule
    providers: *provider_gcp
    depends_on: this.cloud-url-map
    source: "GoogleCloudPlatform/lb-http/google"
    version: "9.2.0"
    inputs:
      name: {{ .variables.project_name }}
      project: {{ .variables.google_project_id }}
      url_map: {{ remoteState "this.cloud-url-map.url_map_self_link" }}
      create_url_map: false
      ssl: false
      backends:
        default:
          protocol: "HTTP"
          port: 80
          port_name: "http"
          timeout_sec: 10
          enable_cdn: false
          groups: [] 
          health_check:
            request_path: "/"
            port: 80
          log_config:
            enable: true
            sample_rate: 1.0
          iap_config:
            enable: false
  -
    name: outputs
    type: printer
    depends_on: this.cloud-storage
    outputs:
      websiteUrl: http://{{ remoteState "this.cloud-lb.external_ip" }}
EOF
```

<details>
  <summary>Click to expand explanation of the Stack Template</summary>

 <h4>1. Provider Definition (_p)</h4> <br>

This section uses a YAML anchor, defining the cloud provider and location for the resources in the stack. For this case, GCS is the chosen provider, and the location is dynamically retrieved from the variables:

```yaml
_p: &provider_gcp
- google:
    project: {{ .variables.google_project_id }}
    region: {{ .variables.google_cloud_region }}
```

<h4>2. Units</h4> <br>

The units section is where the real action is. Each unit is a self-contained "piece" of infrastructure, typically associated with a particular Terraform module or a direct cloud resource. <br>

&nbsp;

<h5>Cloud Storage Unit</h5> <br>

This unit leverages the `github.com/terraform-google-modules/terraform-google-cloud-storage` module to provision an Google Storage Bucket. Inputs for the module, such as the bucket name and project, are filled using variables passed into the Stack.

```yaml
name: cloud-storage
type: tfmodule
providers: *provider_gcp
source: "github.com/terraform-google-modules/terraform-google-cloud-storage.git?ref=v4.0.1"
  inputs:
    names:
      - {{ .variables.name }}
    randomize_suffix: true
    project_id: {{ .variables.google_project_id }}
    location: {{ .variables.google_cloud_bucket_location }}
    set_viewer_roles: true
    viewers:
      - allUsers
    website:
      main_page_suffix: "index.html"
      not_found_page: "index.html"
```

<h5>Cloud Bucket Object Unit</h5> <br>

Upon creating the storage bucket, this unit takes the role of establishing a web-page object inside it. This action is carried out using a module storage bucket object module specifically designed for blob creation. A standout feature is the remoteState function, which dynamically extracts the name of the Storage Bucket name produced by the preceding unit:

```yaml
name: cloud-bucket-object
type: tfmodule
providers: *provider_gcp
depends_on: this.cloud-storage
source: "bootlabstech/cloud-storage-bucket-object/google"
version: "1.0.1"
inputs:
  bucket_name: {{ remoteState "this.cloud-storage.name" }}
  object_name: "index.html"
  object_content: |
    {{- .variables.web_page_content | nindent 8 }}
```

<h5>Cloud URL Map Unit</h5> <br>

This unit create google_compute_url_map and google_compute_backend_bucket in order to supply it to cloud-lb unit. A standout feature is the remoteState function, which dynamically extracts the name of the Storage Bucket name produced by Cloud Storage unit:

```yaml
name: cloud-url-map
type: tfmodule
providers: *provider_gcp
depends_on: this.cloud-storage
source: "github.com/shalb/terraform-gcs-bucket-backend.git?ref=0.0.1"
inputs:
  name: {{ .variables.project_name }}
  bucket_name: {{ remoteState "this.cloud-storage.name" }}
```

<h5>Cloud Load Balancer Unit</h5> <br>

This unit create google load balancer. A standout feature is the remoteState function, which dynamically extracts the name of the URL Map URI produced by Cloud URL Map unit:

```yaml
name: cloud-lb
type: tfmodule
providers: *provider_gcp
depends_on: this.cloud-url-map
source: "GoogleCloudPlatform/lb-http/google"
version: "9.2.0"
inputs:
  name: {{ .variables.project_name }}
  project: {{ .variables.google_project_id }}
  url_map: {{ remoteState "this.cloud-url-map.url_map_self_link" }}
  create_url_map: false
  ssl: false
  backends:
    default:
      protocol: "HTTP"
      port: 80
      port_name: "http"
      timeout_sec: 10
      enable_cdn: false
      groups: [] 
      health_check:
        request_path: "/"
        port: 80
      log_config:
        enable: true
        sample_rate: 1.0
      iap_config:
        enable: false
```

<h5>Outputs Unit</h5> <br>

Lastly, this unit is designed to provide outputs, allowing users to view certain results of the Stack execution. For this template, it provides the website URL of the hosted website exposed by load balancer.

```yaml
name: outputs
type: printer
depends_on: this.cloud-storage
outputs:
  websiteUrl: http://{{ remoteState "this.cloud-lb.external_ip" }}
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
  <p>Now hosted on GCS!</p>
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

## Clean up

To remove the cluster with created resources run the command:

```bash
cdev destroy
```

## More Examples

In the [Examples](https://docs.cluster.dev/examples-overview/) section you will find ready-to-use advanced Cluster.dev samples that will help you bootstrap more complex cloud infrastructures with Helm and Terraform compositions:

- [More Advanced example with GKE](https://docs.cluster.dev/examples-gcp-gke/)
