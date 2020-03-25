resource "random_password" "argocd_pass" {
  length  = 16
  special = false
}

resource "null_resource" "bcrypted_password" {
  triggers = {
    result = bcrypt(random_password.argocd_pass.result)
  }
  lifecycle {
    ignore_changes = all
  }
}

provider "helm" {
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
  version    = "2.0.0"
  namespace  = "argocd"

  values = [
    file("./values.yaml")
  ]
  depends_on = [
    null_resource.kubeconfig_update,
  ]
  set {
    name  = "server.certificate.domain"
    value = var.argo_domain
  }
  set {
    name  = "server.ingress.annotations.\"cluster\\.dev/domain\""
    value = var.argo_domain
  }
  set {
    name  = "server.ingress.hosts[0]"
    value = var.argo_domain
  }
  set {
    name  = "server.ingress.tls[0].hosts[0]"
    value = var.argo_domain
  }
  set {
    name  = "server.config.url"
    value = "https://${var.argo_domain}"
  }
  set {
    name  = "configs.secret.argocdServerAdminPassword"
    value = null_resource.bcrypted_password.triggers.result
  }
  set {
    name  = "installCRDs"
    value = "false"
  }
}

output "argocd_url" {
  value = "https://${var.argo_domain}"
}

output "argocd_user" {
  value = "admin"
}

output "argocd_pass" {
  value = random_password.argocd_pass.result
}
