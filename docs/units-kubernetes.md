# Kubernetes Unit

!!! Info
    The unit is deprecated and will be removed soon. Please use the k8s-manifest unit instead.

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

* `source` - *string*, *required*. Path to Kubernetes manifest that will be converted into a representation of kubernetes-alpha provider. **Source file will be rendered with the stack template, and also allows to use functions `remoteState` and `insertYAML`**.

* `kubeconfig` - *string*, *required*. Path to the kubeconfig file, which is relative to the directory where the unit was executed.
* `provider_version` - *string*, *optional*. Version of terraform kubernetes-alpha provider to use. Default - latest. See [terraform kubernetes-alpha provider](https://registry.terraform.io/providers/hashicorp/kubernetes-alpha/latest) 
