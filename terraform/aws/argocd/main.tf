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