provider "helm" {
  install_tiller = true
  version = "~> 0.10.0"
  service_account = kubernetes_service_account.tiller.metadata.0.name
  namespace = kubernetes_service_account.tiller.metadata.0.namespace
  tiller_image = "gcr.io/kubernetes-helm/tiller:v2.14.1"
}

module "argocd" {
  source =  "./module"
  kubeconfig = "~/.kube/config"
}

resource "null_resource" "kubeconfig_update" {
  triggers = {
    policy_sha1 = "${sha1(file("~/.kube/config"))}"
  }
}

resource "kubernetes_service_account" "tiller" {
  metadata {
    name = "terraform-tiller"
    namespace = "kube-system"
  }
  automount_service_account_token = true

  depends_on = [
    null_resource.kubeconfig_update
  ]

}

resource "kubernetes_cluster_role_binding" "tiller" {
  metadata {
    name = "terraform-tiller"
  }

  role_ref {
    kind = "ClusterRole"
    name = "cluster-admin"
    api_group = "rbac.authorization.k8s.io"
  }

  subject {
    kind = "ServiceAccount"
    name = "terraform-tiller"

    api_group = ""
    namespace = "kube-system"
  }

  depends_on = [
    null_resource.kubeconfig_update
  ]

}