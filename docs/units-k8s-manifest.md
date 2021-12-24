# K8s-manifest Unit

Applies Kubernetes resources from manifests. 

Example:

```yaml
- name: kubectl-test2
  type: k8s-manifest
  namespace: default
  path: ./manifests/
  apply_template: true
  kubeconfig: {{ output "this.kubeconfig.kubeconfig_path" }}
  kubectl_opts: "--wait=true"
```

## Options

* `namespace` - *optional*. Corresponds to `kubectl -n`

* `path` - *required*, *string*. Depending on the variable value, the unit will apply all manifests recursively (if the path points out to a directory) or a file (if points out to the file).

    The `path` value can also be a URL:

    ```yaml
    - name: kubectl-test2
      type: k8s-manifest
      namespace: default
      path: https://git.io/vPieo
      kubeconfig: {{ output "this.kubeconfig.kubeconfig_path" }}
    ```

* `apply_template` - *bool*. By default is set to `true`. See the [Templating usage](#templating-usage) subsection below. 

* `kubeconfig` - *optional*. Specifies the path to a kubeconfig file. 

* `kubectl_opts` - *optional*. Lists additional arguments of the `kubectl` command.   

## Templating usage

As manifests are part of a stack template, they also maintain templating options. Specifying the `apply_template` option enables you to use templating in all Kubernetes manifests located with the specified `path`. 







