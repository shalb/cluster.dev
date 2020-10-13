# Install Operator Lifecycle Manager
locals {
  olm-version   = "0.15.1"
  url           = "https://github.com/operator-framework/operator-lifecycle-manager/releases/download"
  olm-namespace = "olm"
}

resource "null_resource" "olm_install" {
  provisioner "local-exec" {
    # Deploy CRD's, wait them become available and deploy olm
    command = "kubectl apply --kubeconfig ${var.config_path} -f ${local.url}/${local.olm-version}/crds.yaml && kubectl apply --kubeconfig ${var.config_path} -f ${local.url}/${local.olm-version}/olm.yaml && for resource in catalog-operator olm-operator packageserver; do kubectl --kubeconfig ${var.config_path} rollout status -n ${local.olm-namespace} -w deploy $resource; done"
  }
  provisioner "local-exec" {
    when = destroy
    # No sense to delete it
    command = "exit 0"
  }
  depends_on = [
    null_resource.kubeconfig_update,
  ]
}
