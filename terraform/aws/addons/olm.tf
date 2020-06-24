# Install Operator Lifecycle Manager
locals {
  olm-version = "0.15.1"
  url         = "https://github.com/operator-framework/operator-lifecycle-manager/releases/download"
  olm-namespace   = "olm"
}

resource "null_resource" "olm_install" {
  provisioner "local-exec" {
    # Deploy CRD's, wait them become available and deploy olm
    command = "kubectl apply -f ${local.url}/${local.olm-version}/crds.yaml && for resource in catalog-operator olm-operator packageserver; do kubectl rollout status -n ${local.olm-namespace} -w deploy $resource; done && kubectl apply -f ${local.url}/${local.olm-version}/olm.yaml"
  }
  provisioner "local-exec" {
    when    = destroy
    command = "kubectl delete -f ${local.url}/${local.olm-version}/olm.yaml >/dev/null 2&>1"
  }
  depends_on = [
    null_resource.kubeconfig_update,
  ]
}

