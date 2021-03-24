# Template Development

## Basics

A template is a directory, either local or located in a Git repo that contains template config files. cdev reads all ./*.yaml files from the directory (non-recursively), renders a template with the project's data, parse the yaml file and loads modules. Modules may contain reference to other files that are required for work. These files should be located inside the current directory (template context). As some of the files will also be rendered with the project's data, you can use Go-templates in them. For more details please see [modules configuration](#modules) below.

Template represents a yaml structure with an array of different invocation modules. Common view:

```yaml
modules:
  - module1
  - module2
  - module3
  ...
```

Template can utilize all kinds of Go-templates and Sprig functions (similar to Helm). Along with that it is enhanced with functions like insertYAML that could pass yaml blocks directly.

## Functions

1) [Base Go-template language functions](https://golang.org/pkg/text/template/#hdr-Functions).

2) [Sprig functions](https://masterminds.github.io/sprig/).

3) Enhanced functions: all functions described above allow you to modify the template text. Apart from these, some special enhanced functions are available. They cannot be used everywhere. The functions are integrated within the functionality of the program and with the yaml syntax:

* `insertYAML` - pass yaml block as value of target yaml template. **Argument**: data to pass, any value or reference to block. **Allowed use**: only as full yaml value, in module `inputs`. Example:

Source yaml:

```yaml
values:
  node_groups:
    - name: ng1
      min_size: 1
      max_size: 5
    - name: ng2
      max_size: 2
      type: spot
```

Target yaml template:

```yaml
modules:
  - name: k3s
    type: terraform
    node_groups: {{ insertYAML .values.node_groups }}
```

Rendered template:

```yaml
modules:
  - name: k3s
    type: terraform
    node_groups:
    - name: ng1
      min_size: 1
      max_size: 5
    - name: ng2
      max_size: 2
      type: spot
```

* `remoteState` - is used for passing data between modules and infrastructures. **Argument**: string, path to remote state consisting of 3 parts separated by a dot: `"infra_name.module_name.output_name"`. Since the name of the infrastructure is unknown inside the template, you can use "this" instead:`"this.module_name.output_name"`. **Allowed use**: as yaml value, only in module `inputs`.

## Modules

All modules described below have a common format and common fields. Base example:

```yaml
  - name: k3s
    type: terraform
    depends_on:
      - this.module1_name
      - this.module2_name
#   depends_on: this.module1_name # is allowed to use string for single, or list for multiple dependencies
    pre_hook:
      command: "echo pre_hook"
      # script: "./scripts/hook.sh"
      on_apply: true
      on_destroy: false
      on_plan: false
    post_hook:
      # command: "echo post_hook"
      script: "./scripts/hook.sh"
      on_apply: true
      on_destroy: false
      on_plan: false
```

* `name` - module name. *Required*.

* `type` - module type. One of: `terraform`, `helm`, `kubernetes`, `printer`. See below.

* `depends_on` - *string* or *list of strings*. One or multiple module dependencies in the format "infra_name.module_name". Since the name of the infrastructure is unknown inside the template, you can use "this" instead:`"this.module_name.output_name"`.

* `pre_hook` and `post_hook` blocks: describe the shell commands to be executed before and after the module, respectively. The commands will be executed in the same context as the actions of the module. Environment variables are common to the shell commands, the pre_hook and post_hook scripts, and the module execution. You can export a variable in the pre_hook and it will be available in the post_hook or in the module.

    * `command` - *string*. Shell command in text format. Will be executed in bash -c "command". Can be used if the "script" option is not used. One of `command` or `script` is required.

    * `script` - *string* path to shell script file which is relative to template directory. Can be used if the "command" option is not used. One of `command` or `script` is required.

    * `on_apply` *bool*, *optional* turn off/on when module applying. **Default: "true"**.

    * `on_destroy` - *bool*, *optional* turn off/on when module destroying. **Default: "false"**.

    * `on_plan` - *bool*, *optional* turn off/on when module plan executing. **Default: "false"**.

### Terraform module

Describes direct Terraform module invocation.

Example:

```yaml
modules:
  - name: vpc
    type: terraform
    version: "2.77.0"
    source: terraform-aws-modules/vpc/aws
    inputs:
      name: {{ .name }}
      azs: {{ insertYAML .variables.azs }}
      vpc_id: {{ .variables.vpc_id }}
```

In addition to common options the following are available:

* `source` - *string*, *required*. Terraform module [source](https://www.terraform.io/docs/language/modules/syntax.html#source). **It is not allowed to use local folders in source!**

* `version` - *string*, *optional*. Module [version](https://www.terraform.io/docs/language/modules/syntax.html#version).

* `inputs` - *map of any*, *required*. A map that corresponds to [input variables](https://www.terraform.io/docs/language/values/variables.html) defined by the module. This block allows to use functions `remoteState` and `insertYAML`.

### Helm module

Describes [Terraform Helm provider](https://registry.terraform.io/providers/hashicorp/helm/latest/docs) invocation.

Example:

```yaml
modules:
  - name: argocd
    type: helm
    source:
      repository: "https://argoproj.github.io/argo-helm"
      chart: "argo-cd"
      version: "2.11.0"
    kubeconfig: ../kubeconfig
    depends_on: this.k3s
    pre_hook:
      script: ./scripts/get_kubeconfig.sh ./kubeconfig
      on_destroy: true
      on_plan: true
    additional_options:
      namespace: "argocd"
      create_namespace: true
    inputs:
      global.image.tag: v1.8.3
      service.type: LoadBalancer
```

In addition to common options the following are available:

* `source` - *map*, *required*. Block describes Helm chart source.

  * `chart`, `repository`, `version` - correspond to options with the same name from helm_release resource. See [chart](https://registry.terraform.io/providers/hashicorp/helm/latest/docs/resources/release#chart), [repository](https://registry.terraform.io/providers/hashicorp/helm/latest/docs/resources/release#repository) and [version](https://registry.terraform.io/providers/hashicorp/helm/latest/docs/resources/release#version).

  * `kubeconfig` - *string*, *required*. Path to the kubeconfig file which is relative to the directory where the module was executed.

  * `additional_options` - *map of any*, *optional*. Corresponds to [Terraform helm_release resource options](https://registry.terraform.io/providers/hashicorp/helm/latest/docs/resources/release#argument-reference). Will be passed as is.

  * `inputs` - *map of any*, *optional*. A map that represents [Terraform helm_release sets](https://registry.terraform.io/providers/hashicorp/helm/latest/docs/resources/release#set). This block allows to use functions `remoteState` and `insertYAML`. For example:

   ```yaml
    inputs:
      global.image.tag: v1.8.3
      service.type: LoadBalancer
    ```

    corresponds to:

    ```hcl
    set {
      name = "global.image.tag"
      value = "v1.8.3"
    }
    set  {
      name = "service.type"
      value = "LoadBalancer"
    }
    ```

### Kubernetes module

Describes [Terraform kubernetes-alpha provider](https://github.com/hashicorp/terraform-provider-kubernetes-alpha) invocation.

Example:

```yaml
modules:
  - name: argocd_apps
    type: kubernetes
    source: ./argocd-apps/app1.yaml
    kubeconfig: ../kubeconfig
    depends_on: this.argocd
```

* `source` - *string*, *required*. Path to Kubernetes manifest that will be converted into a representation of kubernetes-alpha provider. **Source file will be rendered with the template, and also allows to use the functions `remoteState` and `insertYAML`**.

* `kubeconfig` - *string*, *required*. Path to the kubeconfig file which is relative to the directory where the module was executed.

### Printer module

The module is mainly used to see the outputs of other modules in the console logs.

Example:

```yaml
modules:
  - name: print_outputs
    type: printer
    inputs:
      cluster_name: {{ .name }}
      worker_iam_role_arn: {{ remoteState "this.eks.worker_iam_role_arn" }}
```

* `inputs` - *any*, *required* - a map that represents data to be printed in the log. The block **allows to use the functions `remoteState` and `insertYAML`**.
