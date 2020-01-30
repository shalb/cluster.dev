module "argocd" { 
  source =  "../module"
  kubeconfig_hash = var.kubeconfig_hash
}