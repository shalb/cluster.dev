# Product Design

Cluster.dev helps you manage Cloud Native Infrastructures with a simple declarative manifests - InfraTemplates.

InfraTemplate could contain different technology patterns, like Terraform modules, Shell scripts, Kubernetes manifests, Helm charts, Kustomize and ArgoCD applications.

## Base Concept

```bash
# Common Infrastructure Project structure
[Project]
          -> /templates
             /infrastructures
             /scripts
             /terraform
             /kubernetes
             /any_outher_technology
```

### Infrastructure Project

Project represents the single scope under it infrastructures are stored and reconciled. The dependencies between different infrastructures could be used under the project scope.  
Project could host a global variables that could be accessed to template target infrastructure.

### Infrastructure Template

Represents yaml structure that includes different invocation modules: Terraform modules, Shell scripts, Kubernetes kubectl, Helm and Kustomize.

Template could utilize all kind of go-template and sprig functions (similar to Helm). Along that it is enhanced with functions like `insertYAML` that could pass yaml blocks directly.

```yaml
# Sample infrastructure template that deploys k8s cluster in AWS
# Define template variables
{{- $gitUrl := "github.com/shalb/cluster.dev//terraform/aws" }}
{{- $branch := "new-reconciler" }}
modules:
# Invoke Terraform (type:terraform) module
- name: route53
    type: terraform
    source: {{ $gitUrl }}/route53?ref={{ $branch }}
    inputs:
      region: {{ .variables.region }}
      cluster_name: {{ .name }}
      cluster_domain: {{ .variables.domain }}
      zone_delegation: {{ if eq .variables.domain "cluster.dev" }}true{{ else }}false{{ end }}
  - name: vpc
    type: terraform
    source: {{ $gitUrl }}/vpc?ref={{ $branch }}
    inputs:
      region: {{ .variables.region }}
      cluster_name: {{ .name }}
      availability_zones: {{ insertYAML .variables.azs }}
  - name: minikube
    type: terraform
    source: {{ $gitUrl }}/minikube?ref={{ $branch }}
    inputs:
      region: {{ .variables.region }}
      cluster_name: {{ .name }}
      aws_instance_type: {{ .variables.instance_type }}
      hosted_zone: {{ .name }}.{{ .variables.domain }}
      aws_subnet_id: {{ remoteState "this.vpc.public_subnets[0]" }}
      bucket_name: {{ .variables.bucket }}
# Execute shell script during execution to set config
  - name: get_kubeconfig
    type: shell
    depends_on: this.minikube
    command: "aws s3 cp s3://{{ .variables.bucket }}/kubeconfig_{{ .name }} ../kubeconfig_{{ .name }} && echo set_output_kubeconfig_path=../kubeconfig{{ .name }} "
# Invoke next Terraform module
  - name: addons
    type: terraform
    source: {{ $gitUrl }}/addons?ref={{ $branch }}
    pre_hook: this.get_kubeconfig
    inputs:
      region: {{ .variables.region }}
      cluster_name: {{ .name }}
      kubeconfig: {{ output "this.get_kubeconfig.kubeconfig_path" }}
      cluster_cloud_domain: {{ .name }}.{{ .variables.domain }}
      dns_zone_id: {{ remoteState "this.route53.zone_id" }}
```

### Infrastructure Variables

The values that would be used by template to render the resulting infrastructure.

```yaml
name: infra-dev2
template: "templates/infra.tmpl"
kind: infrastructure
backend: aws
variables:
  bucket: new-cluster-dev
  region: eu-central-1
  organization: "shalb"
  domain: cluster.dev
  instance_type: "t3.medium"
  env: "dev-test1"
  azs:
    - "eu-central-1a"
    - "eu-central-1b"
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

