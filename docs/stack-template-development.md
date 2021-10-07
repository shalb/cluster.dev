# Stack Template Development

## Basics

A stack template is a yaml file, which tells cdev which units to run and how. It is a core cdev resource that makes for its flexibility. Stack templates use Go template language to allow you customise and select the units you want to run.

The stack template's config files are stored within the stack template directory, which could be located either locally or in a Git repo. cdev reads all _./*.yaml files from the directory (non-recursively), renders a stack template with the project's data, parse the yaml file and loads units - the most primitive elements of a stack template. 

Units are building blocks that stack templates are made of. It could be anything — a Terraform module, Helm you want to install or a Bash script that you want to run. Units can be remote or stored in the same repo with other cdev code. Units may contain reference to other files that are required for work. These files should be located inside the current directory (stack template's context). As some of the files will also be rendered with the project's data, you can use Go templates in them. For more details please see [units configuration](#units) below.

A stack template represents a yaml structure with an array of different invocation units. Common view:

```yaml
units:
  - unit1
  - unit2
  - unit3
  ...
```

Stack templates can utilize all kinds of Go templates and Sprig functions (similar to Helm). Along with that it is enhanced with functions like insertYAML that could pass yaml blocks directly.

## Functions

1) [Base Go template language functions](https://golang.org/pkg/text/template/#hdr-Functions).

2) [Sprig functions](https://masterminds.github.io/sprig/).

3) Enhanced functions: all functions described above allow you to modify the text of a stack template. Apart from these, some special enhanced functions are available. They cannot be used everywhere. The functions are integrated within the functionality of the program and with the yaml syntax:

* `insertYAML` - pass yaml block as value of target yaml template. **Argument**: data to pass, any value or reference to block. **Allowed use**: only as full yaml value, in unit `inputs`. Example:

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
units:
  - name: k3s
    type: terraform
    node_groups: {{ insertYAML .values.node_groups }}
```

Rendered stack template:

```yaml
units:
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

* `remoteState` - is used for passing data across units and stacks, can be used in pre/post hooks. **Argument**: string, path to remote state consisting of 3 parts separated by a dot: `"stack_name.unit_name.output_name"`. Since the name of the stack is unknown inside the stack template, you can use "this" instead:`"this.unit_name.output_name"`. **Allowed use**: 

    * all units types: in `inputs`;

    * all units types: in units pre/post hooks;

    * in Kubernetes modules: in Kubernetes manifests.

* `cidrSubnet` - calculates a subnet address within given IP network address prefix. Same as [Terraform function](https://www.terraform.io/docs/language/functions/cidrsubnet.html). Example:

Source:
```bash
{{ cidrSubnet "172.16.0.0/12" 4 2 }}
```
Rendered:
```bash
172.18.0.0/16
```

## Units

All units described below have a common format and common fields. Base example:

```yaml
  - name: k3s
    type: terraform
    depends_on:
      - this.unit1_name
      - this.unit2_name
#   depends_on: this.unit1_name # is allowed to use string for single, or list for multiple dependencies
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

* `name` - unit name. *Required*.

* `type` - unit type. One of: `terraform`, `helm`, `kubernetes`, `printer`. See below.

* `depends_on` - *string* or *list of strings*. One or multiple unit dependencies in the format "stack_name.unit_name". Since the name of the stack is unknown inside the stack template, you can use "this" instead:`"this.unit_name.output_name"`.

* `pre_hook` and `post_hook` blocks: describe the shell commands to be executed before and after the unit, respectively. The commands will be executed in the same context as the actions of the unit. Environment variables are common to the shell commands, the pre_hook and post_hook scripts, and the unit execution. You can export a variable in the pre_hook and it will be available in the post_hook or in the unit.

    * `command` - *string*. Shell command in text format. Will be executed in bash -c "command". Can be used if the "script" option is not used. One of `command` or `script` is required.

    * `script` - *string*. Path to shell script file which is relative to template directory. Can be used if the "command" option is not used. One of `command` or `script` is required.

    * `on_apply` *bool*, *optional*. Turn off/on when module applying. **Default: "true"**.

    * `on_destroy` - *bool*, *optional*. Turn off/on when module destroying. **Default: "false"**.

    * `on_plan` - *bool*, *optional*. Turn off/on when module plan executing. **Default: "false"**.

### Terraform module

Describes direct Terraform module invocation.

Example:

```yaml
units:
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

### Helm unit

Describes [Terraform Helm provider](https://registry.terraform.io/providers/hashicorp/helm/latest/docs) invocation.

Example:

```yaml
units:
  - name: argocd
    type: helm
    source:
      repository: "https://argoproj.github.io/argo-helm"
      chart: "argo-cd"
      version: "2.11.0"
    pre_hook:
      command: *getKubeconfig
      on_destroy: true
    kubeconfig: ./kubeconfig_{{ .name }}
    depends_on: this.cert-manager-issuer
    additional_options:
      namespace: "argocd"
      create_namespace: true
    values:
      - file: ./argo/values.yaml
        apply_template: true
    inputs:
      global.image.tag: v1.8.3
```

In addition to common options the following are available:

* `source` - *map*, *required*. This block describes Helm chart source.

  * `chart`, `repository`, `version` - correspond to options with the same name from helm_release resource. See [chart](https://registry.terraform.io/providers/hashicorp/helm/latest/docs/resources/release#chart), [repository](https://registry.terraform.io/providers/hashicorp/helm/latest/docs/resources/release#repository) and [version](https://registry.terraform.io/providers/hashicorp/helm/latest/docs/resources/release#version).

* `kubeconfig` - *string*, *required*. Path to the kubeconfig file which is relative to the directory where the unit was executed.
* `provider_version` - *string*, *optional*. Version of terraform helm provider to use. Default - latest. See [terraform helm provider](https://registry.terraform.io/providers/hashicorp/helm/latest)  

* `additional_options` - *map of any*, *optional*. Corresponds to [Terraform helm_release resource options](https://registry.terraform.io/providers/hashicorp/helm/latest/docs/resources/release#argument-reference). Will be passed as is.

* `values` - *array*, *optional*. List of values files in raw yaml to be passed to Helm. Values will be merged, in order, as Helm does with multiple -f options.

    * `file` - *string*, *required*. Path to the values file.

    * `apply_template` - *bool*, *optional*. Defines whether a template should be applied to the values file. By default is set to `true`. 

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

### Kubernetes unit

Describes [Terraform kubernetes-alpha provider](https://github.com/hashicorp/terraform-provider-kubernetes-alpha) invocation.

Example:

```yaml
units:
  - name: argocd_apps
    type: kubernetes
    provider_version: "0.2.1"
    source: ./argocd-apps/app1.yaml
    kubeconfig: ../kubeconfig
    depends_on: this.argocd
```

* `source` - *string*, *required*. Path to Kubernetes manifest that will be converted into a representation of kubernetes-alpha provider. **Source file will be rendered with the stack template, and also allows to use the functions `remoteState` and `insertYAML`**.

* `kubeconfig` - *string*, *required*. Path to the kubeconfig file which is relative to the directory where the unit was executed.
* `provider_version` - *string*, *optional*. Version of terraform kubernetes-alpha provider to use. Default - latest. See [terraform kubernetes-alpha provider](https://registry.terraform.io/providers/hashicorp/kubernetes-alpha/latest) 

### Printer unit

The unit is mainly used to see the outputs of other units in the console logs.

Example:

```yaml
units:
  - name: print_outputs
    type: printer
    inputs:
      cluster_name: {{ .name }}
      worker_iam_role_arn: {{ remoteState "this.eks.worker_iam_role_arn" }}
```

* `inputs` - *any*, *required* - a map that represents data to be printed in the log. The block **allows to use the functions `remoteState` and `insertYAML`**.


