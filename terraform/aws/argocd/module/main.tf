provider "helm" {
  install_tiller = true
  version = "~> 0.10.0"
  service_account = kubernetes_service_account.tiller.metadata.0.name
  namespace = kubernetes_service_account.tiller.metadata.0.namespace
  tiller_image = "gcr.io/kubernetes-helm/tiller:v2.14.1"
}

data "helm_repository" "argo" {
  name = "argo"
  url  = "https://argoproj.github.io/argo-helm"
}

resource "helm_release" "argo-cd" {
  name       = "argo-cd-${kubeconfig_hash}"
  repository = data.helm_repository.argo.metadata[0].name
  chart      = "argo-cd"
  version    = "1.6.3"
  namespace  = "argocd"

  values = [
    "${file("values.yaml")}"
  ]
}

provider "kubernetes" {
}

resource "kubernetes_service_account" "tiller" {
  metadata {
    name = "terraform-tiller-${kubeconfig_hash}"
    namespace = "kube-system"
  }

  automount_service_account_token = true
}

resource "kubernetes_cluster_role_binding" "tiller" {
  metadata {
    name = "terraform-tiller-${kubeconfig_hash}"
  }

  role_ref {
    kind = "ClusterRole"
    name = "cluster-admin"
    api_group = "rbac.authorization.k8s.io"
  }

  subject {
    kind = "ServiceAccount"
    name = "terraform-tiller-${kubeconfig_hash}"

    api_group = ""
    namespace = "kube-system"
  }

}

