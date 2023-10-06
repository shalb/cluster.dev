# **Getting Started with Cluster.dev on Azure Cloud**

This guide will walk you through the steps to deploy your first project with Cluster.dev on Azure Cloud.

```text
                          +-------------------------+
                          | Project.yaml            |
                          |  - location             |
                          +------------+------------+
                                       |
                                       |
                          +------------v------------+
                          | Stack.yaml              |
                          |  - storage_account_name |
                          |  - location             |
                          |  - file_content         |
                          +------------+------------+
                                       |
                                       |
+--------------------------------------v----------------------------------------+
| StackTemplate: azure-static-website                                           |
|                                                                               |
|  +---------------------+     +---------------------+     +-----------------+  |
|  | resource-group      |     | storage-account     |     | web-page-blob   |  |
|  | type: tfmodule      |     | type: tfmodule      |     | type: tfmodule  |  |
|  | inputs:             |     | inputs:             |     | inputs:         |  |
|  |  location           |     | storage_account_name|     |  file_content   |  |
|  |  resource_group_name|     |                     |     |                 |  |
|  +---------------------+     +----------^----------+     +--------^--------+  |
|        |                       | resource-group           | storage-account   |
|        |                       | name & location          | name              |
|        |                       | via remoteState          | via remoteState   |
+--------|-----------------------|--------------------------|-------------------+
         |                       |                          |
         v                       v                          v
Azure Resource Group    Azure Storage Account      Azure Blob (in $web container)
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

- **Azure CLI**:

  ```bash
  curl -sL https://aka.ms/InstallAzureCLIDeb | sudo bash
  az --version
  ```

- **Cluster.dev client**:

  ```bash
  curl -fsSL https://raw.githubusercontent.com/shalb/cluster.dev/master/scripts/get_cdev.sh | sh
  cdev --version
  ```

## Authentication

Before using the Azure CLI, you'll need to authenticate:

  ```bash
   az login --use-device-code
  ```

Follow the prompt to sign in.

## Creating an Azure Blob Storage for Storing State

First, create a resource group:

  ```bash
  az group create --name cdevResourceGroup --location EastUS
  ```

Then, create a storage account:

  ```bash
  az storage account create --name cdevstates --resource-group cdevResourceGroup --location EastUS --sku Standard_LRS
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
backend: azure-backend
variables:
  organization: cluster.dev
  location: eastus
  state_storage_account_name: cdevstates
EOF
```

### Backend Configuration (`backend.yaml`)

This specifies where Cluster.dev will store its own state and the Terraform states for any infrastructure it provisions or manages. Given the backend type as S3, it's clear that AWS is the chosen cloud provider.

```bash
cat <<EOF > backend.yaml
name: azure-backend
kind: Backend
provider: local ## to be changed to azurerm
spec:
  resource_group_name: cdevResourceGroup
  storage_account_name: {{ .project.variables.state_storage_account_name }}
  container_name: tfstate
EOF
```

### Stack Configuration (`stack.yaml`)

*   This represents a distinct set of infrastructure resources to be provisioned.
*   It references a local template (in this case, the previously provided stack template) to know what resources to create.
*   Variables specified in this file will be passed to the Terraform modules called in the template.
*   The content variable here is especially useful; it dynamically populates the content of an S3 bucket object (a webpage in this case) using the local `index.html` file.

```bash
cat <<EOF > stack.yaml
name: az-blob-website
template: ./template/
kind: Stack
backend: azure-backend
variables:
  storage_account_name: "tmpldevtest"
  resource_group_name: "demo-resource-group"
  location: {{ .project.variables.location }}
  file_content: |
    {{- readFile "./files/index.html" | nindent 4 }}
EOF
```

### Stack Template (`template.yaml`)

The `StackTemplate` serves as a pivotal object within Cluster.dev. It lays out the actual infrastructure components you intend to provision using Terraform modules and resources. Essentially, it determines how your cloud resources should be laid out and interconnected.

```bash
mkdir template
cat <<EOF > template/template.yaml
_p: &provider_azurerm
- azurerm:
    features:
      resource_group:
        prevent_deletion_if_contains_resources: false

_globals: &global_settings
  default_region: "region1"
  regions:
    region1: {{ .variables.location }}
  prefixes: ["dev"]
  random_length: 4
  passthrough: false
  use_slug: false
  inherit_tags: false

_version: &module_version 5.7.5

name: azure-static-website
kind: StackTemplate
units:
  -
    name: resource-group
    type: tfmodule
    providers: *provider_azurerm
    source: aztfmod/caf/azurerm//modules/resource_group
    version: *module_version
    inputs:
      global_settings: *global_settings
      resource_group_name: {{ .variables.resource_group_name }}
      settings:
        region: "region1"
  -
    name: storage-account
    type: tfmodule
    providers: *provider_azurerm
    source: aztfmod/caf/azurerm//modules/storage_account
    version: *module_version
    inputs:
      base_tags: false
      global_settings: *global_settings
      client_config:
        key: demo
      resource_group:
        name: {{ remoteState "this.resource-group.name" }}
        location: {{ remoteState "this.resource-group.location" }}
      storage_account:
        name: {{ .variables.storage_account_name }}
        account_kind: "StorageV2"
        account_tier: "Standard"
        static_website:
          index_document: "index.html"
          error_404_document: "error.html"
      var_folder_path: "./"
  -
    name: web-page-blob
    type: tfmodule
    providers: *provider_azurerm
    source: aztfmod/caf/azurerm//modules/storage_account/blob
    version: *module_version
    inputs:
      settings:
        name: "index.html"
        content_type: "text/html"
        source_content: |
          {{- .variables.file_content | nindent 12 }}
      storage_account_name: {{ remoteState "this.storage-account.name" }}
      storage_container_name: "$web"
      var_folder_path: "./"
  -
    name: outputs
    type: printer
    depends_on: this.web-page-blob
    outputs:
      websiteUrl: https://{{ remoteState "this.storage-account.primary_web_host" }}
EOF
```

<details>
  <summary>Click to expand explanation of the Stack Template</summary>

 <h4>1. Provider Definition (_p)</h4> <br>

This section uses a YAML anchor, defining the cloud provider and location for the resources in the stack. For this case, Azure is the chosen provider, and the location is dynamically retrieved from the variables:

```yaml
_p: &provider_azurerm
- azurerm:
    features:
      resource_group:
        prevent_deletion_if_contains_resources: false
```

<h4>2. Units</h4> <br>

The units section is where the real action is. Each unit is a self-contained "piece" of infrastructure, typically associated with a particular Terraform module or a direct cloud resource. <br>

&nbsp;  

<h5>Storage Account Unit</h5> <br>

This unit leverages the `aztfmod/caf/azurerm//modules/storage_account` module to provision an Azure Blob Storage account. Inputs for the module, such as the storage account name, are filled using variables passed into the Stack.

```yaml
name: storage-account
type: tfmodule
providers: *provider_azurerm
source: aztfmod/caf/azurerm//modules/storage_account
inputs:
  name: {{ .variables.storage_account_name }}
  ...
```

<h5>Web-page Object Unit</h5> <br>

Upon creating the storage account, this unit takes the role of establishing a web-page object inside it. This action is carried out using a sub-module from the storage account module specifically designed for blob creation. A standout feature is the remoteState function, which dynamically extracts the name of the Azure Storage account produced by the preceding unit:

```yaml
name: web-page-blob
type: tfmodule
providers: *provider_azurerm
source: aztfmod/caf/azurerm//modules/storage_account/blob
inputs:
  storage_account_name: {{ remoteState "this.storage-account.name" }}
  ...
```

<h5>Outputs Unit</h5> <br>

Lastly, this unit is designed to provide outputs, allowing users to view certain results of the Stack execution. For this template, it provides the website URL of the hosted S3 website.

```yaml
name: outputs
type: printer
depends_on: this.web-page-blob
outputs:
  websiteUrl: https://{{ remoteState "this.storage-account.primary_web_host" }}
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
  <p>Now hosted on Azure!</p>
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

<a href="https://asciinema.org/a/8j5WthKY9exW74ldotwCqWPcE" target="_blank"><img src="https://asciinema.org/a/8j5WthKY9exW74ldotwCqWPcE.svg" /></a>

## Clean up

To remove the cluster with created resources run the command:

```bash
cdev destroy
```
