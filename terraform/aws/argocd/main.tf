resource "kubernetes_service_account" "tiller" {
  metadata {
    name = "tiller"
    namespace = "kube-system"
  }
}
resource "kubernetes_cluster_role_binding" "tiller" {
  metadata {
        name = "tiller"
  }
  subject {
    api_group = "rbac.authorization.k8s.io"
    kind      = "User"
    name      = "system:serviceaccount:kube-system:tiller"
  }

  role_ref {
    api_group = "rbac.authorization.k8s.io"
    kind  = "ClusterRole"
    name = "cluster-admin"
  }
  depends_on = [kubernetes_service_account.tiller]
}

provider "helm" {
    tiller_image = "gcr.io/kubernetes-helm/tiller:v2.12.3"
    install_tiller = true
    service_account = "tiller"
    namespace = "kube-system"
    debug = true
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