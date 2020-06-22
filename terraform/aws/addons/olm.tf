# Install Operator Lifecycle Manager
locals {
  olm-version = "0.15.1"
  url         = "https://github.com/operator-framework/operator-lifecycle-manager/releases/download"
  olm-namespace   = "olm"
}

resource "null_resource" "olm_install" {
  provisioner "local-exec" {
    command = "kubectl apply -f ${local.url}/${local.olm-version}/crds.yaml -f ${local.url}/${local.olm-version}/olm.yaml && kubectl rollout status -w deployment/olm-operator -n ${local.olm-namespace}"
  }
  provisioner "local-exec" {
    when    = destroy
    command = "kubectl delete -f ${local.url}/${local.olm-version}/crds.yaml -f ${local.url}/${local.olm-version}/olm.yaml >/dev/null 2&>1"
  }
  depends_on = [
    null_resource.kubeconfig_update,
  ]
}

