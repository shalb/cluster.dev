# Kubectl Unit

Applies Kubernetes resources from manifests. 

Example:

```yaml
name: kubectl-test2
    type: kubectl
    namespace: default
    path: ./manifests/
    pre_hook:
      command: *getKubeconfig
      on_destroy: true
      on_plan: true
    apply_template: true
    kubeconfig: ./kubeconfig_{{ .name }}
    kubectl_opts: "--wait=true"
```

## Options

* `namespace` - *optional*. Corresponds to `kubectl -n`

* `path` - *required*, *string*. Depending on the variable value, the unit will apply all manifests recursively (if the path points out to a directory) or a file (if points out to the file).

* `pre_hook` - See the description in [Shell unit](https://docs.cluster.dev/units-shell/#options).

* `apply_template` - *bool*. By default is set to `true`. See the [Templating usage](#templating-usage) subsection below. 

* `kubeconfig` - *optional*. Specifies the path to a kubeconfig file.

* `kubectl_opts` - *optional*. Lists additional arguments of the `kubectl` command.   

## Templating usage

As manifests are part of a stack template, they also maintain templating options. Specifying the `apply_template` option enables you to use templating in all Kubernetes manifests located with the specified `path`. 







