# K8s-manifest Unit

!!! Info

    This unit is deprecated. We suggest using the Kubernetes unit instead.

Applies Kubernetes resources from manifests. 

Example:

```yaml
- name: kubectl-test2
  type: k8s-manifest
  namespace: default
  create_namespaces: true
  path: ./manifests/
  apply_template: true
  kubeconfig: {{ output "this.kubeconfig.kubeconfig_path" }}
  kubectl_opts: "--wait=true"
```

## Options

* `force_apply` - *bool*, *optional*. By default is false. If set to true, the unit will be applied when any dependent unit is planned to be changed.

* `namespace` - *optional*. Corresponds to `kubectl -n`.

* `create_namespaces` - *bool*, *optional*. By default is false. If set to true, cdev will create namespaces required for the unit (i.e. the namespaces listed in manifests and the one specified within the `namespace` option), in case they don't exist.

* `path` - *required*, *string*. Indicates the resources that are to be applied: a file (in case of a file path), a directory recursively (in case of a directory path) or URL. In case of URL path the unit will download the resources by the link and then apply them.  

    Example of URL `path`:

    ```yaml
    - name: kubectl-test2
      type: k8s-manifest
      namespace: default
      path: https://git.io/vPieo
      kubeconfig: {{ output "this.kubeconfig.kubeconfig_path" }}
    ```

* `apply_template` - *bool*. By default is set to `true`. See [Templating usage](#templating-usage) below.

* `kubeconfig` - *optional*. Specifies the path to a kubeconfig file. See [How to get kubeconfig](#how-to-get-kubeconfig) subsection below.

* `kubectl_opts` - *optional*. Lists additional arguments of the `kubectl` command.

## Templating usage

As manifests are part of a stack template, they also maintain [templating](https://docs.cluster.dev/templating/) options. Specifying the `apply_template` option enables you to use templating in all Kubernetes manifests located with the specified `path`.

## How to get kubeconfig

There are several ways to get a kubeconfig from a cluster and pass it to the units that require it (for example, `helm`, `K8s-manifest`). The recommended way is to use the `shell` unit with the option `force_apply`. Here is an example of such unit:

```yaml
- name: kubeconfig
  type: shell
  force_apply: true
  depends_on: this.k3s
  apply:
    commands:
      - aws s3 cp s3://{{ .variables.bucket }}/{{ .variables.cluster_name }}/kubeconfig /tmp/kubeconfig_{{ .variables.cluster_name }}
      - echo "kubeconfig_base64=$(cat /tmp/kubeconfig_{{ .variables.cluster_name }} | base64 -w 0)"
      - echo "kubeconfig_path=/tmp/kubeconfig_{{ .variables.cluster_name }}"
  outputs:
    type: separator
    separator: "="
```

In the example above, the `K3s` unit (the one referred to) will deploy a Kubernetes cluster in AWS and place a kubeconfig file in S3 bucket. The `kubeconfig` unit will download the kubeconfig from the storage and place it within the /tmp directory. 

The kubeconfig can then be passed as an output to other units:

```yaml
- name: cert-manager-issuer
  type: k8s-manifest
  path: ./cert-manager/issuer.yaml
  kubeconfig: {{ output "this.kubeconfig.kubeconfig_path" }}
```

An alternative (but not recommended) way is to create a yaml hook in a stack template that would take the required set of commands:

```yaml
_: &getKubeconfig "rm -f ../kubeconfig_{{ .name }}; aws eks --region {{ .variables.region }} update-kubeconfig --name {{ .name }} --kubeconfig ../kubeconfig_{{ .name }}"
```

and execute it with a pre-hook in each unit:

```yaml
- name: cert-manager-issuer
  type: kubernetes
  source: ./cert-manager/
  provider_version: "0.6.0"
  config_path: ../kubeconfig_{{ .name }}
  depends_on: this.cert-manager
  pre_hook:
    command: *getKubeconfig
    on_destroy: true
    on_plan: true
```





