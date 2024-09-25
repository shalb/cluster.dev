# Kubernetes Unit

The unit employs the [Terraform Kubernetes provider](https://registry.terraform.io/providers/hashicorp/kubernetes/latest/docs) to interact with the resources supported by Kubernetes. Unlike the Terraform Kubernetes provider, it supports direct input of Kubernetes manifests, allowing the use of `remoteState` and Cluster.dev templating in manifest files.

The unit automatically converts Kubernetes manifests into Terraform resources by parsing the `apiVersion` and `kind` fields. Check the Kubernetes provider [documentation](https://registry.terraform.io/providers/hashicorp/kubernetes/latest/docs) for a list of supported resources. Custom or unsupported resources are converted into the universal resource [kubernetes_manifest](https://registry.terraform.io/providers/hashicorp/kubernetes/latest/docs/resources/manifest).

!!! tip

    To deploy to Kubernetes using Terraform, use this unit to automatically convert YAML files into ready-to-use HCL code.

## Example usage

```yaml
units:
  - name: argocd_apps
    type: kubernetes
    source: ./argocd-apps/app1.yaml
    kubeconfig: ../kubeconfig
    depends_on: this.argocd
```

## Options

* `force_apply` - *bool*, *optional*. By default is false. If set to true, the unit will be applied when any dependent unit is changed.

* `source` - *string*, *required*. Similar to option `path` in the `k8s-manifest` unit. Indicates the resources that are to be applied: a file (in case of a file path), a directory recursively (in case of a directory path) or URL. In case of URL path, the unit will download the resources by the link and then apply them. **Source file will be rendered with the stack template, and also allows to use functions `remoteState` and `insertYAML`**.

* `kubeconfig` - *string*, *required*. Path to the kubeconfig file, which is relative to the directory where the unit was executed.

* `apply_template` - *bool*. By default is set to `true`. Enables applying [templating](https://docs.cluster.dev/templating/) to all Kubernetes manifests located in the specified `path`. 

* [`provider_conf`](#provider_conf) - configuration block that describes authorization in Kubernetes. Supports the same arguments as the [Terraform Kubernetes provider](https://registry.terraform.io/providers/hashicorp/kubernetes/latest/docs#argument-reference). It is allowed to use the `remoteState` function and Cluster.dev templates within the block. For details see below.

### `provider_conf`

Example usage:

  ```yaml
    name: cert-manager-issuer
    type: kubernetes
    depends_on: this.cert-manager
    source: ./deployment.yaml
    provider_conf:
      host: k8s.example.com
      username: "user"
      password: "secretPassword"
  ```

* `host` - *optional*. The hostname (in form of URI) of the Kubernetes API. Can be sourced from `KUBE_HOST`.

* `username` - *optional*. The username to use for HTTP basic authentication when accessing the Kubernetes API. Can be sourced from `KUBE_USER`.

* `password` - *optional*. The password to use for HTTP basic authentication when accessing the Kubernetes API. Can be sourced from `KUBE_PASSWORD`.

* `insecure` - *optional*. Whether the server should be accessed without verifying the TLS certificate. Can be sourced from `KUBE_INSECURE`. Defaults to `false`.

* `tls_server_name` - *optional*. Server name passed to the server for SNI and is used in the client to check server certificates against. Can be sourced from `KUBE_TLS_SERVER_NAME`.

* `client_certificate` - *optional*. PEM-encoded client certificate for TLS authentication. Can be sourced from `KUBE_CLIENT_CERT_DATA`.

* `client_key` - *optional*. PEM-encoded client certificate key for TLS authentication. Can be sourced from `KUBE_CLIENT_KEY_DATA`.

* `client_ca_certificate` - *optional*. PEM-encoded root certificates bundle for TLS authentication. Can be sourced from `KUBE_CLUSTER_CA_CERT_DATA`.

* `config_path` - *optional*. A path to a kube config file. Can be sourced from `KUBE_CONFIG_PATH`.

* `config_paths` - *optional*. A list of paths to the kube config files. Can be sourced from `KUBE_CONFIG_PATHS`.

* `config_context` - *optional*. Context to choose from the config file. Can be sourced from `KUBE_CTX`.

* `config_context_auth_info` - *optional*. Authentication info context of the kube config (name of the kubeconfig user, `--user` flag in `kubectl`). Can be sourced from `KUBE_CTX_AUTH_INFO`.

* `config_context_cluster` - *optional*. Cluster context of the kube config (name of the kubeconfig cluster, `--cluster` flag in `kubectl`). Can be sourced from `KUBE_CTX_CLUSTER`.

* `token` - *optional*. Token of your service account. Can be sourced from `KUBE_TOKEN`

* `proxy_url` - *optional*. URL to the proxy to be used for all API requests. URLs with "http", "https", and "socks5" schemes are supported. Can be sourced from `KUBE_PROXY_URL`.

* `exec` - *optional*. Configuration block to use an [exec-based credential plugin](https://kubernetes.io/docs/reference/access-authn-authz/authentication/#client-go-credential-plugins), e.g. call an external command to receive user credentials.

    * `api_version` - *required*. API version to use when decoding the ExecCredentials resource, e.g. `client.authentication.k8s.io/v1beta1`.

    * `command` - *required*. Command to execute.

    * `args` - *optional*. List of arguments to pass when executing the plugin.

    * `env` - *optional*. Map of environment variables to set when executing the plugin.

* `ignore_annotations` - *optional*.  List of Kubernetes metadata annotations to ignore across all resources handled by this provider for situations where external systems are managing certain resource annotations. This option does not affect annotations within a template block. Each item is a regular expression.

* `ignore_labels` - *optional*. List of Kubernetes metadata labels to ignore across all resources handled by this provider for situations where external systems are managing certain resource labels. This option does not affect annotations within a template block. Each item is a regular expression.



