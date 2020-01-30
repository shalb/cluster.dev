variable "kubeconfig_hash" {
  description = "Hash from kubeconfig to re-create Tiller and Argo when it changed"
}

module "argocd" { 
  source =  "./module"
  kubeconfig_hash = var.kubeconfig_hash
}