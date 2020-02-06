provider "helm" {
  version = "~> 1.0.0"
}

# Check for kubeconfig and update resources when cluster gets re-created.
resource "null_resource" "kubeconfig_update" {
  triggers = {
    policy_sha1 = "${sha1(file("~/.kube/config"))}"
  }
}

# Deploy ArgoCD 
resource "kubernetes_namespace" "argocd" {
  metadata {
    name = "argocd"
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
  version    = "1.7.1"
  namespace  = "argocd"

  values = [
    "${file("values.yaml")}"
  ]
  depends_on = [
    null_resource.kubeconfig_update,
      ]
}