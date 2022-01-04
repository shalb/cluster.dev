# K8s-manifest Unit

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

* `kubeconfig` - *optional*. Specifies the path to a kubeconfig file. See the [How to get kubeconfig](#how-to-get-kubeconfig) subsection below.

* `kubectl_opts` - *optional*. Lists additional arguments of the `kubectl` command. 

## Templating usage

As manifests are part of a stack template, they also maintain templating options. Specifying the `apply_template` option enables you to use templating in all Kubernetes manifests located with the specified `path`. 

## How to get kubeconfig

Instead of using `pre_hook`, Cluster.dev has a dedicated unit to get and pass a kubeconfig file: 

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

In the example above, the `K3s` unit (the one referred to) will deploy a Kubernetes cluster in AWS and place a kubeconfig file in S3 bucket. The `kubeconfig` unit will download the kubeconfig from the storage and place it within the /tmp directory. Then it can be passed to other units:

```yaml
- name: cert-manager-issuer
  type: k8s-manifest
  path: ./cert-manager/issuer.yaml
  kubeconfig: {{ output "this.kubeconfig.kubeconfig_path" }}
```

!!! Info
    Please make sure to set `force_apply: true`. Specifying the option is compulsory in such cases.






