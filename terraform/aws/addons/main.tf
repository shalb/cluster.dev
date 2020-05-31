# Check for kubeconfig and update resources when cluster gets re-created.
resource "null_resource" "kubeconfig_update" {
  triggers = {
    policy_sha1 = "${sha1(file(var.config_path))}"
  }
}

#######################
# Deploy nginx-ingress
#######################

resource "null_resource" "nginx_ingress_install" {
  triggers = {
    config_contents = filemd5(var.config_path)
  }
  provisioner "local-exec" {
    command = "kubectl apply --kubeconfig ${var.config_path} -f 'https://raw.githubusercontent.com/shalb/terraform-aws-minikube/master/addons/ingress.yaml'"
  }
  provisioner "local-exec" {
    when = destroy
    # destroy ingress object to remove created Load Balancer
    command = "kubectl delete --kubeconfig ${var.config_path} -f 'https://raw.githubusercontent.com/shalb/terraform-aws-minikube/master/addons/ingress.yaml'"
  }
}

######################
# Deploy ArgoCD
######################
locals {
  argocd_domain = "argocd.${var.cluster_name}.${var.cluster_cloud_domain}"
}

# Generate and bcrypt password
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

# Create ArgoCD namespace
resource "kubernetes_namespace" "argocd" {
  metadata {
    name = "argocd"
  }
}

# Deploy ArgoCD with Helm provider
resource "helm_release" "argo-cd" {
  name       = "argo-cd"
  repository = "https://argoproj.github.io/argo-helm"
  chart      = "argo-cd"
  version    = "2.0.0"
  namespace  = "argocd"

  values = [
    file("templates/argocd-values.yaml")
  ]
  depends_on = [
    null_resource.kubeconfig_update,
    helm_release.cert_manager,
  ]
  set {
    name  = "server.certificate.domain"
    value = local.argocd_domain
  }
  set {
    name  = "server.ingress.annotations.\"cluster\\.dev/domain\""
    value = local.argocd_domain
  }
  set {
    name  = "server.ingress.hosts[0]"
    value = local.argocd_domain
  }
  set {
    name  = "server.ingress.tls[0].hosts[0]"
    value = local.argocd_domain
  }
  set {
    name  = "server.config.url"
    value = "https://${local.argocd_domain}"
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
