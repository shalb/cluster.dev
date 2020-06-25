#####################
# Deploy ExternalSecrets
#####################
# Create Namespace for ExternalSecrets
resource "kubernetes_namespace" "external-secrets" {
  metadata {
    name = "external-secrets"
  }
}

resource "helm_release" "external-secrets" {
  name       = "external-secrets"
  repository = "https://godaddy.github.io/kubernetes-external-secrets/"
  chart      = "kubernetes-external-secrets"
  version    = "4.0.0"
  namespace  = "external-secrets"
  depends_on = [
    null_resource.kubeconfig_update,
    kubernetes_namespace.external-secrets,
  ]
  set {
    name  = "aws.region"
    value = var.region
  }
}