resource "null_resource" "kubeconfig_update" {
  triggers = {
    policy_sha1 = "${sha1(file(var.kubeconfig))}"
  }
}

data "helm_repository" "argo" {
  name = "argo"
  url  = "https://argoproj.github.io/argo-helm"
}

resource "helm_release" "argo-cd" {
  name       = "argo-cd"
  repository = data.helm_repository.argo.metadata[0].name
  chart      = "argo-cd"
  version    = "1.6.3"
  namespace  = "argocd"

  values = [
    "${file("${path.module}/values.yaml")}"
  ]
  depends_on = [
    null_resource.kubeconfig_update,
      ]
}