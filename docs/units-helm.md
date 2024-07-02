# Helm Unit

Describes [Terraform Helm provider](https://registry.terraform.io/providers/hashicorp/helm/latest/docs) invocation.

## Example usage

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

## Options 

* `force_apply` - *bool*, *optional*. By default is false. If set to true, the unit will be applied when any dependent unit is changed.

* `source` - *map*, *required*. This block describes Helm chart source.

* `chart`, `repository`, `version` - correspond to options with the same name from helm_release resource. See [chart](https://registry.terraform.io/providers/hashicorp/helm/latest/docs/resources/release#chart), [repository](https://registry.terraform.io/providers/hashicorp/helm/latest/docs/resources/release#repository) and [version](https://registry.terraform.io/providers/hashicorp/helm/latest/docs/resources/release#version).

* `kubeconfig` - *string*, *required*. Path to the kubeconfig file which is relative to the directory where the unit was executed.
* `provider_version` - *string*, *optional*. Version of Terraform Helm provider to use. Default - latest. See [terraform helm provider](https://registry.terraform.io/providers/hashicorp/helm/latest)  

* `additional_options` - *map of any*, *optional*. Corresponds to [Terraform helm_release resource options](https://registry.terraform.io/providers/hashicorp/helm/latest/docs/resources/release#argument-reference). Will be passed as is.

* [`values`](#values) - *array*, *optional*. List of values (file name or values data) to be passed to Helm. Values will be merged, in order, as Helm does with multiple -f options. For details see below.

* `inputs` - *map of any*, *optional*. A map that represents [Terraform helm_release sets](https://registry.terraform.io/providers/hashicorp/helm/latest/docs/resources/release#set). This block allows to use functions `remoteState` and `insertYAML`.  
For example:

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

* [`provider_conf`](#provider_conf) - configuration block that describes authorization in Kubernetes. Supports the same arguments as the [Terraform Kubernetes provider](https://registry.terraform.io/providers/hashicorp/kubernetes/latest/docs#argument-reference). It is allowed to use the `remoteState` function and Cluster.dev templates within the block. For details see below.

### `values`

* `set` - *map of any*, *required one of set/file*. Set of Helm values. This option allows you to transfer the value of the Helm chart without saving it to a file.

* `file` - *string*, *required one of set/file*. Path to the values file.

* `apply_template` - *bool*, *optional*. Defines whether a template should be applied to the values file. By default is set to `true`. Used only with `file` option.

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


