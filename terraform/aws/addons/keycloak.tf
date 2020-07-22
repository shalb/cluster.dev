# Deploy keycloak-operator and keycloak itself
locals {
  cluster_issuer     = "letsencrypt-prod"
  keycloak_namespace = "keycloak-operator"
  keycloak_domain    = "keycloak.${var.cluster_name}.${var.cluster_cloud_domain}"
}

# Deploy Keycloak Operator
resource "null_resource" "keycloak-operator_install" {
  provisioner "local-exec" {
    command = "kubectl apply --kubeconfig ${var.config_path} -f templates/keycloak-operator.yaml && until kubectl --kubeconfig ${var.config_path} -n ${local.keycloak_namespace} wait --for condition=established --timeout=60s crd/keycloaks.keycloak.org >/dev/null 2>&1; do sleep 1; echo waiting keycloak crd; done"
  }
  provisioner "local-exec" {
    when    = destroy
    command = "exit 0"
  }
  depends_on = [
    null_resource.kubeconfig_update,
    null_resource.olm_install
  ]
}

# Deploy Keycloak
data "template_file" "keycloak" {
  template = file("templates/keycloak.yaml")
  vars = {
    domain         = local.keycloak_domain
    cluster-issuer = local.cluster_issuer
    namespace      = local.keycloak_namespace
  }
}

resource "null_resource" "keycloak_install" {
  provisioner "local-exec" {
    command = "kubectl apply --kubeconfig ${var.config_path} -f -<<EOF\n${data.template_file.keycloak.rendered}\nEOF"
  }
  provisioner "local-exec" {
    command = "until kubectl --kubeconfig ${var.config_path} -n ${local.keycloak_namespace} get secrets/credential-keycloak >/dev/null 2>&1; do sleep 1; echo 'Waiting for secrets/credential-keycloak become available'; done"
  }
  provisioner "local-exec" {
    when    = destroy
    command = "exit 0"
  }
  depends_on = [
    null_resource.kubeconfig_update,
    null_resource.keycloak-operator_install
  ]
}

# Get credentials from generated config map
data "kubernetes_secret" "keycloak_credentials" {
  metadata {
    name      = "credential-keycloak"
    namespace = local.keycloak_namespace
  }
  depends_on = [
    null_resource.keycloak_install,
    null_resource.keycloak-operator_install
  ]
}
