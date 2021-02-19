# Cluster.dev: Cloud infrastructures management tool.

Cluster.dev helps you manage Cloud Native Infrastructures with a simple declarative manifests - Infra Templates.

So you can describe whole infrastructure an deploy it with single tool.

InfraTemplate could contain different technology patterns, like Terraform modules, Shell scripts, Kubernetes manifests, Helm charts, Kustomize and ArgoCD/Flux applications, OPA policies etc..

You can store, test, and distribute your infrastructure pattern as a complete versioned set of technologies.

**Table of contents:**

1) [Concept](##Concept)
2) [Install](##Install)
    - [Prerequisites](###Prerequisites)
    - [Download from release](###Download-from-release)
    - [Build from source](###Build-from-source)
3) [Quick start]()
    - [AWS]()
    - [Google Cloud]()
    - [Azure]()
    - [DigitalOcean]()
4) [Command arguments and options]()
5) [Project configuration]()
    - [Project]()
    - [Infrastructures]()
    - [Backends]()
    - [Secrets]()
      - [SOPS]()
      - [Amazon secret manager]()
6) [Template configuration]()
    - [Basics]()
    - [Functions]()
    - [Modules]()
      - [Terraform]()
      - [Helm]()
      - [Kubernetes]()
      - [Printer]()
7) [Generators]()
    - [Project]()
    - [Secret]()

## Install
### Prerequisites
#### Terraform
Cdev client uses the terraform binary. The required terraform version is ~13 or higher. Refer to the [Terraform installation instructions](https://www.terraform.io/downloads.html) to install terraform.

Terraform installation example for linux amd64:
```bash
curl -O https://releases.hashicorp.com/terraform/0.14.7/terraform_0.14.7_linux_amd64.zip
unzip terraform_0.14.7_linux_amd64.zip
mv terraform /usr/local/bin/
```
#### SOPS
For **creating** and **editing** SOPS secrets, cdev uses sops binary. 

But sops binary in not required for decrypting and using sops secrets. All cdev reconcilation processes (build, plan, apply) is not required sops, so no need to install it for pipelines.

See [SOPS installation instructions](https://github.com/mozilla/sops#download) on official repo.

Also see [Secrets section]() in this documentation.

### Download from release


Binaries and packages of the latest stable release are available at [Releases page](https://github.com/shalb/cluster.dev/releases). This documentation is suitable for **v0.4.0 or higher**
Cdev client installation example for linux amd64:
```bash
wget https://github.com/shalb/cluster.dev/releases/download/v0.4.0-rc1/cluster.dev_v0.4.0-rc1_linux_amd64.tgz
tar -xzvf cluster.dev_v0.4.0-rc1_linux_amd64.tgz -C /usr/local/bin

cdev --help
```

### Build from source

Go version 16 or higher is required. [Golang installation instructions](https://golang.org/doc/install)

To build cluster.dev client from src:
1)  Clone cluster.dev git repo:
```bash
git clone --depth 1 --branch v0.4.0-rc1 https://github.com/shalb/cluster.dev/
```
2) Build binary: 
```bash
cd cluster.dev/ && make
```
3) Check cdev and move binary to bin folder: 
```bash
./bin/cdev --help
mv ./bin/cdev /usr/local/bin/
```

## Concept

```bash
# Common Infrastructure Project Structure
[Project in Git Repo]
  project.yaml           # (Required) Global variables and settings
  [filename].yaml        # (Required at least one) Different project's objects in yaml format (infrastructure, backend etc). 
                         # Se details in configuration reference.
  /templates             # Pre-defined infra patterns. See details in template configuration reference.
    aws-kubernetes.yaml
    cloudflare-dns.yaml
    do-mysql.yaml

    /files               # Some files, used in templates.
      deployment.yaml
      config.cfg
```

### Infrastructure Reconcilation

```bash
# Single command reconcile the whole project
cdev apply
```

Would:

 1. Decode all required secrets.
 2. Template infrastructure variables with global project variables and secrets.
 3. Pull and diff project state and build a dependency graph.
 4. Invoke all required modules in parralel manner.
    ex: `sops decode`, `terraform apply`, `helm install`, etc..

### Demo Video

video placeholder

## Base Components

### Infrastructure Project

Project represents the single scope for infrastructures unde what they are stored and reconciled. The dependencies between different infrastructures could be used under the project scope.
Project could host a global variables that could be accessed to template target infrastructure.

```yaml
# Sample project
name: yolo
kind: project
variables:
  # global vars could be used across envs
  organization: shalb
  region: eu-central-1
  state_bucket_name: cluster-dev-state-bucket
```

### Infrastructure Template

Represents yaml structure that includes different invocation modules: Terraform modules, Shell scripts, Kubernetes kubectl, Helm and Kustomize.

Template could utilize all kind of go-template and sprig functions (similar to Helm). Along that it is enhanced with functions like `insertYAML` that could pass yaml blocks directly.

```yaml
# Sample infrastructure template
# It deploys k8s cluster in AWS, install ArgoCD on it
# and apply nginx-ingress as ArgoCD application.
modules:
  # Create DNS record with Terraform
  - name: route53
    type: terraform
    source: github.com/shalb/cluster.dev//terraform/aws/route53?ref=v0.5.0
    inputs:
      region: {{ .variables.region }}
      cluster_name: {{ .name }}
      cluster_domain: {{ .variables.domain }}
      zone_delegation: {{ if eq .variables.domain "cluster.dev" }}true{{ else }}false{{ end }}
  # If VPC is not provided create a AWS VPC with Antons module
  {{- if not .variables.vpc_id }}
  - name: vpc
    type: terraform
    source: terraform-aws-modules/vpc/aws
    version: "v1.73.0"
    inputs:
      name: {{ .name }}
      azs: {{ insertYAML .variables.azs }}
      vpc_id: {{ .variables.vpc_id }}
  {{- end }}
  # Deploy k3s cluster with Terraform
  - name: k3s
    type: terraform
    source: github.com/shalb/terraform-aws-k3s.git?ref=v0.1.2
    inputs:
      cluster_name: {{ .name }}-1
      extra_args:
        - "--disable traefik"
      domain: {{ remoteState "this.route53.domain" }}
      kubeconfig_filename: "../kubeconfig"
      k3s_version: "1.19.3+k3s1"
      {{- if not .variables.vpc_id }}
      public_subnets: {{ remoteState "this.vpc.public_subnets" }}
      {{- end }}
      {{- if .variables.vpc_id }}
      public_subnets: {{ insertYAML .variables.public_subnets }}
      {{- end }}
      key_name: {{ remoteState "this.aws_key_pair.this_key_pair_key_name" }}
      region: {{ .variables.region }}
      s3_bucket: {{ .variables.bucket }}
      master_node_count: 1
      worker_node_groups: {{ insertYAML .variables.worker_node_groups  }}
  # Deploy ArgoCD to provisioned K3s cluster with Helm
  - name: argocd
    type: helm
    source:
      repository: "https://argoproj.github.io/argo-helm"
      chart: "argo-cd"
      version: "2.11.0"
    pre_hook:
      command: export KUBECONFIG=../kubeconfig || aws s3 cp s3://{{ .variables.bucket }}/{{ .name }}-1/kubeconfig ../kubeconfig
      on_destroy: true
    kubeconfig: ../kubeconfig
    depends_on: this.k3s
    additional_options:
      namespace: "argocd"
      create_namespace: true
    inputs:
      global.image.tag: v1.8.3
      service.type: LoadBalancer
      server.certificate.domain: argocd.{{ .name }}.{{ .variables.domain }}
      server.ingress.enabled: true
      server.ingress.hosts[0]: argocd.{{ .name }}.{{ .variables.domain }}
      server.ingress.tls[0].hosts[0]: argocd.{{ .name }}.{{ .variables.domain }}
      server.config.url: https://argocd.{{ .name }}.{{ .variables.domain }}
      configs.secret.argocdServerAdminPassword: {{ .variables.argocd_secret }}
  # Deploy ArgoCD applications as Kubernetes raw manifest
  - name: argocd_apps
    type: kubernetes
    source: ./argocd-apps/
    kubeconfig: ../kubeconfig
    depends_on: this.argocd
```

### Infrastructure Variables

The values that would be used by template to render the resulting infrastructure.

```yaml
# Define Backend to store Terraform states and other stuff
name: aws-backend
kind: backend
provider: s3
spec:
  bucket: {{ .project.variables.state_bucket_name }}
  region: {{ .project.variables.region }}
---
# Define infrastructure itself
name: k3s-infra
template: "templates/aws-k3s.yaml" # use template above
kind: infrastructure
backend: aws-backend
variables:
  bucket: {{ .project.variables.state_bucket_name }}
  region: {{ .project.variables.region }}
  organization: {{ .project.variables.organization }}
  domain: cluster.dev
  instance_type: "t3.medium"
  vpc_id: "vpc-5ecf1234"
  public_subnets:
    - "subnet-d775f0bd"
    - "subnet-6696651a"
  env: "dev"
  argocd_secret: {{ .secrets.infra-dev.argocd_secret }}
  azs:
    - {{ .project.variables.region }}a
  public_key_name: voa
  worker_node_groups:
    - name: "node_pool"
      min_size: 0
      max_size: 1
      instance_type: "t3.medium"
```

### Infrastructure Secrets

There are to way use secrets:
#### sops
Secrets are encoded/decoded with [sops](https://github.com/mozilla/sops) utility which could support AWS KMS, GCP KMS, Azure Key Vault and PGP keys.
How to use: 
1. Use console client cdev to create new secret from scratch:
```bash
cdev secret create sops my_local_secret
```

2. Edit secret and set secret data in `encrypted_data:` section.

3. Use references to secret's data it infrastructure template. 
```yaml
---
name: my-infra
template: "templates/k8s.yaml"
kind: infrastructure
backend: aws-backend
variables:
  region: {{ .secret.my_local_secret.some_key }}
....
```

#### aws ssm
cdev client could use aws ssm as secret storage. 

How to use: 
1. create new secret in aws ssm manager using aws cli or web console. Both data format raw and JSON structure are supported.

2. Use console client cdev to create new secret from scratch:
```bash
cdev secret create aws_ssm my_ssm_secret
```

3. Edit secret and set correct region and ssm_secret name in spec.

4. Use references to secret's data it infrastructure template. 
```yaml
---
name: my-infra
template: "templates/k8s.yaml"
kind: infrastructure
backend: aws-backend
variables:
  region: {{ .secret.my_ssm_secret.some_key }} # if secret is raw data use {{ .secret.my_ssm_secret }}
....
```

To list and edit any secret use commands:
```bash
cdev secret ls
```
and 
```bash
cdev secret edit secret_name
```

## Reconcilation Modules

### Terraform Modules

Reconciler supports direct Terraform module invocation. So you can use public or private modules just to define required dependencies.
The following module definition:

```yaml
- name: minikube
  type: terraform
  source: https://github.com/shalb/cluster.dev/terraform/aws/minikube?ref=1.1
  inputs:
    region: {{ .variables.region }}
    cluster_name: {{ .name }}
    aws_instance_type: {{ .variables.instance_type }}
    hosted_zone: {{ .name }}.{{ .variables.domain }}
    aws_subnet_id: {{ remoteState "this.vpc.public_subnets[0]" }}
    bucket_name: {{ .variables.bucket }}
```

Would:

1. Download a module to local filesystem from remote location.
2. Render all input variables to json values and place them to module folder.
3. Generate a separate remote state file definition to module folder.
4. Check module dependencies across whole project, and generate a remote state definitions to obtain values from other modules.
5. Set the invocation priority based on dependency graph.
6. Apply Terraform inside the module folder and save outputs.

### Shell Modules

Reconciler supports a shell script execution during a different phases of reconcilation.
An example of execution for AWS cli command in the middle of the different Terraform modules apply:

```yaml
  - name: get_kubeconfig
    type: shell
    depends_on: this.minikube
    command: "aws s3 cp s3://{{ .variables.bucket }}/kubeconfig_{{ .name }} ../kubeconfig_{{ .name }} && echo set_output_kubeconfig_path=../kubeconfig{{ .name }} "
```

