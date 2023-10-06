# **Getting Started with Kubernetes and Helm**

This guide will walk you through the steps to deploy a WordPress application along with a MySQL database on a Kubernetes cluster using StackTemplates with [Helm units](https://docs.cluster.dev/units-helm/).

```text
                          +-------------------------+
                          | Stack.yaml              |
                          |  - domain               |
                          |  - kubeconfig_path      |
                          +------------+------------+
                                       |
                                       |
+--------------------------------------v---------------------------------+
| StackTemplate: wordpress                                               |
|                                                                        |
|  +---------------------+               +---------------------+         |
|  | mysql-wp-pass-user  |-------------->| mysql-wordpress     |         |
|  | type: tfmodule      |               | type: helm          |         |
|  | output:             |               | inputs:             |         |
|  |  generated password |               |  kubeconfig         |         |
|  |                     |               |  values (from mysql.yaml)     |
|  +---------------------+               +----------|----------+         |
|                                                   |                    |
|                                                   v                    |
|                                           MySQL Deployment             |
|                                                   |                    |
|  +---------------------+               +----------|----------+         |
|  | wp-pass             |-------------->| wordpress           |         |
|  | type: tfmodule      |               | type: helm          |         |
|  | output:             |               | inputs:             |         |
|  |  generated password |               |  kubeconfig         |         |
|  |                     |               |  values (from wordpress.yaml) |
|  +---------------------+               +----------|----------+         |
|                                                   |                    |
|                                                   v                    |
|                                           WordPress Deployment         |
|                                                                        |
|  +---------------------+                                               |
|  | outputs             |                                               |
|  | type: printer       |                                               |
|  | outputs:            |                                               |
|  |  wordpress_url      |                                               |
|  +---------------------+                                               |
|            |                                                           |
+------------|-----------------------------------------------------------+
             |
             v
      wordpress_url Output

```

## **Prerequisites**

1. A running Kubernetes cluster.
2. Your domain name (for this tutorial, we'll use `example.com` as a placeholder).
3. The `kubeconfig` file for your Kubernetes cluster.

## Setting Up Your Project

!!! tip
You can clone example files from repo:
```bash
git clone https://github.com/shalb/cdev-examples
cd cdev-examples/helm/wordpress/
```

### Project Configuration (`project.yaml`)

- Defines the overarching project settings. All subsequent stack configurations will inherit and can override these settings.
- It points to aws-backend as the backend, meaning that the Cluster.dev state for resources defined in this project will be stored in the S3 bucket specified in `backend.yaml`.
- Project-level variables are defined here and can be referenced in other configurations.

```bash
cat <<EOF > project.yaml
name: wordpress-demo
kind: Project
backend: aws-backend
variables:
  region: eu-central-1
  state_bucket_name: cdev-states
EOF
```

### Backend Configuration (`backend.yaml`)

This specifies where Cluster.dev will store its own state and the Terraform states for any infrastructure it provisions or manages. In this example the AWS s3 is used, but you can choose [any other provider](https://docs.cluster.dev/structure-backend/).

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

### Setting Up the Stack File (`stack.yaml`)

- This represents a high level of infrastructure pattern configuration.
- It references a local template to know what resources to create.
- Variables specified in this file will be passed to the Terraform modules and Helm charts called in the template.

Replace placeholders in `stack.yaml` with your actual `kubeconfig` path and domain.

```bash
cat <<EOF > stack.yaml
name: wordpress
template: "./template/"
kind: Stack
backend: aws-backend
cliVersion: ">= 0.7.14"
variables:
  kubeconfig_path: "/data/home/voa/projects/cdev-aws-eks/examples/kubeconfig" # Change to your path
  domain: demo.cluster.dev # Change to your domain
EOF
```

### Stack Template (template.yaml)

The StackTemplate serves as a pivotal object within Cluster.dev. It lays out the actual infrastructure components you intend to provision using Terraform modules and resources. Essentially, it determines how your cloud resources should be laid out and interconnected.

```bash
mkdir template
cat <<EOF > template/template.yaml
kind: StackTemplate
name: wordpress
cliVersion: ">= 0.7.15"
units:
## Generate Passwords with Terraform for MySQL and Wordpress
  -
    name: mysql-wp-pass-user
    type: tfmodule
    source: github.com/romanprog/terraform-password?ref=0.0.1
    inputs:
      length: 12
      special: false
  -
    name: wp-pass
    type: tfmodule
    source: github.com/romanprog/terraform-password?ref=0.0.1
    inputs:
      length: 12
      special: false
## Install MySQL and Wordpress with Helm
  -
    name: mysql-wordpress
    type: helm
    kubeconfig: {{ .variables.kubeconfig_path }}
    source:
      repository: "oci://registry-1.docker.io/bitnamicharts"
      chart: "mysql"
      version: "9.9.1"
    additional_options:
      namespace: "wordpress"
      create_namespace: true
    values:
      - file: ./files/mysql.yaml
  -
    name: wordpress
    type: helm
    depends_on: this.mysql-wordpress
    kubeconfig: {{ .variables.kubeconfig_path }}
    source:
      repository: "oci://registry-1.docker.io/bitnamicharts"
      chart: "wordpress"
      version: "16.1.2"
    additional_options:
      namespace: "wordpress"
      create_namespace: true
    values:
      - file: ./files/wordpress.yaml

  - name: outputs
    type: printer
    depends_on: this.wordpress
    outputs:
      wordpress_url: https://wordpress.{{ .variables.domain }}/admin/
      wordpress_user: user
      wordpress_password: {{ remoteState "this.wp-pass.result" }}
EOF
```

As you can see the StackTemplate contains Helm units and they could use inputs from values.yaml files where it is possible to use outputs from other type of units(like tfmodule) or even other stacks. Lets create that values for MySQL and Wordpress:

```bash
mkdir files
cat <<EOF > files/mysql.yaml
fullNameOverride: mysql-wordpress
auth:
  rootPassword: {{ remoteState "this.mysql-wp-pass-user.result" }}
  username: user
  password: {{ remoteState "this.mysql-wp-pass-user.result" }}
EOF
```

```bash
cat <<EOF > files/wordpress.yaml
containerSecurityContext:
  enabled: false
mariadb:
  enabled: false
externalDatabase:
  port: 3306
  user: user
  password: {{ remoteState "this.mysql-wp-pass-user.result" }}
  database: my_database
wordpressPassword: {{ remoteState "this.wp-pass.result" }}
allowOverrideNone: false
ingress:
  enabled: true
  ingressClassName: "nginx"
  pathType: Prefix
  hostname: wordpress.{{ .variables.domain }}
  path: /
  tls: true
  annotations:
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
EOF
```

<details>
  <summary>Click to expand explanation of the Stack Template</summary>

<h4>1. Units</h4> <br>

The units section is a list of infrastructure components that are provisioned sequentially. Each unit has a type, which indicates whether it's a Terraform module (`tfmodule`), a Helm chart (`helm`), or simply outputs (`printer`).

<h5>Password Generation Units</h5> <br>

There are two password generation units which use the Terraform module `github.com/romanprog/terraform-password` to generate random passwords.

```yaml
name: mysql-wp-pass-user
type: tfmodule
source: github.com/romanprog/terraform-password?ref=0.0.1
inputs:
  length: 12
  special: false
```

These units will create passwords with a length of 12 characters without special characters. The outputs of these units (the generated passwords) are used in subsequent units.

<h5> MySQL Helm Chart Unit</h5> <br>

This unit installs the MySQL chart from the `bitnamicharts` Helm repository.

```yaml
name: mysql-wordpress
type: helm
kubeconfig: {{ .variables.kubeconfig_path }}
source:
  repository: "oci://registry-1.docker.io/bitnamicharts"
  chart: "mysql"
  version: "9.9.1"
```

The `kubeconfig` field uses a variable to point to the Kubeconfig file, enabling Helm to deploy to the correct Kubernetes cluster.

<h5>WordPress Helm Chart Unit</h5> <br>

This unit installs the WordPress chart from the same Helm repository as MySQL. It depends on the `mysql-wordpress` unit, ensuring MySQL is installed first.

```yaml
name: wordpress
type: helm
depends_on: this.mysql-wordpress
```

Both Helm units utilize external YAML files (`mysql.yaml` and `wordpress.yaml`) to populate values for the Helm charts. These values files leverage the `remoteState` function to fetch passwords generated by the Terraform modules.

<h5>Outputs Unit</h5> <br>

This unit outputs the URL to access the WordPress site.

```yaml
name: outputs
type: printer
depends_on: this.wordpress
outputs:
  wordpress_url: https://wordpress.{{ .variables.domain }}/admin/
```

It waits for the WordPress Helm unit to complete (`depends_on: this.wordpress`) and then provides the URL.

<h4>2. Variables and Data Flow</h4> <br>

In this stack template:

- The `.variables` placeholders, like `{{ .variables.kubeconfig_path }}` and `{{ .variables.domain }}`, fetch values from the stack's variables.
- The `remoteState` function, such as `{{ remoteState "this.wp-pass.result" }}`, fetches the outputs from previous units. For example, it retrieves the randomly generated password for WordPress.

These mechanisms ensure dynamic configurations based on real-time resource states and user-defined variables. They enable values generated in one unit (e.g., a password from a Terraform module) to be utilized in a subsequent unit (e.g., a Helm deployment).

<h4>3. Additional File (`mysql.yaml` and `wordpress.yaml`) Explanation</h4> <br>

Both files serve as value configurations for their respective Helm charts.

- `mysql.yaml` sets overrides for the MySQL deployment, specifically the authentication details.
- `wordpress.yaml` customizes the WordPress deployment, such as its database settings, ingress configuration, and password.

Both files leverage the `remoteState` function to pull in passwords generated by the Terraform password modules.

In summary, this stack template and its additional files define a robust deployment that sets up a WordPress application with its database, all while dynamically creating and injecting passwords. It showcases the synergy between Terraform for infrastructure provisioning and Helm for Kubernetes-based application deployments.

</details>

## Deploying WordPress and MySQL with cluster.dev

### 1. Planning the Deployment

   ```bash
   cdev plan
   ```

### 2. Applying the StackTemplate

   ```bash
   cdev apply
   ```

Upon executing these commands, WordPress and MySQL will be deployed on your Kubernetes cluster using cluster.dev.

### Example Screen Cast

<a href="https://asciinema.org/a/PKdBskqaUKB2zvURad7qPm469" target="_blank"><img src="https://asciinema.org/a/PKdBskqaUKB2zvURad7qPm469.svg" /></a>

## Clean up

To remove the cluster with created resources run the command:

```bash
cdev destroy
```

## Conclusion

StackTemplates provide a modular approach to deploying applications on Kubernetes. With Helm and StackTemplates, you can efficiently maintain, scale, and manage your deployments. This guide walked you through deploying WordPress and MySQL seamlessly on a Kubernetes cluster using these tools.

## More Examples

In the Examples section you will find ready-to-use advanced Cluster.dev samples that will help you bootstrap more complex cloud infrastructures with Helm and Terraform compositions:

- [Install EKS cluster with Wordpress as separate stack in one project](https://github.com/shalb/cdev-aws-eks/tree/main/examples)
- [Install sample application by templating multiple Kubernetes manifests](https://github.com/shalb/cdev-aws-k3s/tree/main/examples/sample-application-template)
- [Repo with other examples](https://github.com/shalb/cdev-examples)

