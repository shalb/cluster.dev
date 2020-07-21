#####################
# Deploy ExternalDNS
#####################
# Create Namespace for ExternalDNS
resource "kubernetes_namespace" "external-dns" {
  metadata {
    name = "external-dns"
  }
}

resource "helm_release" "external-dns" {
  name       = "external-dns"
  repository = "https://charts.bitnami.com/bitnami"
  chart      = "external-dns"
  version    = "2.20.10"
  namespace  = "external-dns"
  depends_on = [
    null_resource.kubeconfig_update,
    kubernetes_namespace.external-dns,
  ]
  set {
    name  = "rbac.create"
    value = "true"
  }
  set {
    name  = "provider"
    value = "digitalocean"
  }
  set {
    name  = "digitalocean.apiToken"
    value = var.do_token
  }
  set {
    name  = "interval"
    value = "1m"
  }
  set {
    name  = "policy"
    value = "sync"
  }
}
