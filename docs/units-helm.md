# Helm Unit

Describes [Terraform Helm provider](https://registry.terraform.io/providers/hashicorp/helm/latest/docs) invocation.

In the example below we use `helm` unit to deploy Argo CD to a Kubernetes cluster:

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
    kubeconfig: /home/john/kubeconfig
    additional_options:
      namespace: "argocd"
      create_namespace: true
    values:
      - file: ./argo/values.yaml
        apply_template: true
      - set:
          global:
            image:
              tag: "v1.8.3"
      - set: {{ insertYAML .variables.argocd.values }}
    inputs:
      global.image.tag: v1.8.3 # (same as values.set )
```

In addition to common options the following are available:

* `force_apply` - *bool*, *optional*. By default is false. If set to true, the unit will be applied when any dependent unit is changed.

* `source` - *map*, *required*. This block describes Helm chart source.

* `chart`, `repository`, `version` - correspond to options with the same name from helm_release resource. See [chart](https://registry.terraform.io/providers/hashicorp/helm/latest/docs/resources/release#chart), [repository](https://registry.terraform.io/providers/hashicorp/helm/latest/docs/resources/release#repository) and [version](https://registry.terraform.io/providers/hashicorp/helm/latest/docs/resources/release#version).

* `kubeconfig` - *string*, *required*. Path to the kubeconfig file which is relative to the directory where the unit was executed.
* `provider_version` - *string*, *optional*. Version of Terraform Helm provider to use. Default - latest. See [terraform helm provider](https://registry.terraform.io/providers/hashicorp/helm/latest)  

* `additional_options` - *map of any*, *optional*. Corresponds to [Terraform helm_release resource options](https://registry.terraform.io/providers/hashicorp/helm/latest/docs/resources/release#argument-reference). Will be passed as is.

* `values` - *array*, *optional*. List of values (file name or values data) to be passed to Helm. Values will be merged, in order, as Helm does with multiple -f options.

    * `set` - *map of any*, *required one of set/file*. Set of Helm values. This option allows you to transfer the value of the Helm chart without saving it to a file.

    * `file` - *string*, *required one of set/file*. Path to the values file.

    * `apply_template` - *bool*, *optional*. Defines whether a template should be applied to the values file. By default is set to `true`. Used only with `file` option.

* `inputs` - *map of any*, *optional*. A map that represents [Terraform helm_release sets](https://registry.terraform.io/providers/hashicorp/helm/latest/docs/resources/release#set). This block allows to use functions `remoteState` and `insertYAML`. For example:

```yaml
  inputs:
    global.image.tag: v1.8.3
    service.type: LoadBalancer
  ```

Corresponds to:

```yaml
      set {
        name = "global.image.tag"
        value = "v1.8.3"
      }
      set  {
        name = "service.type"
        value = "LoadBalancer"
      }
```
