provider "helm" {
  version        = "~> 0.10"
  install_tiller = true
}

data "helm_repository" "argo" {
  name = "argo"
  url  = "https://argoproj.github.io/argo-helm"
}

resource "helm_release" "argo-cd" {
  name       = "argo-cd"
  repository = data.helm_repository.argo.metadata[0].name
  chart      = "argo"
  version    = "1.6.3"
  namespace  = "argocd"

  values = [
    "${file("values.yaml")}"
  ]
}