provider "helm" {
}
#####################
# Deploy ExternalDNS
#####################

# Check for kubeconfig and update resources when cluster gets re-created.
resource "null_resource" "kubeconfig_update" {
  triggers = {
    policy_sha1 = "${sha1(file(var.config_path))}"
  }
}

# Create Namespace for ExternalDNS
resource "kubernetes_namespace" "external-dns" {
  metadata {
    name = "external-dns"
  }
}

data "helm_repository" "bitnami" {
  name = "bitnami"
  url  = "https://charts.bitnami.com/bitnami"
}

resource "helm_release" "external-dns" {
  name       = "external-dns"
  repository = data.helm_repository.bitnami.metadata[0].name
  chart      = "external-dns"
  version    = "2.20.10"
  namespace  = "external-dns"

  values = [
    file("external-dns-values.yaml")
  ]
  depends_on = [
    null_resource.kubeconfig_update,
    kubernetes_namespace.external-dns,
  ]
  set {
    name  = "aws.region"
    value = var.aws_region
  }
}

######################
# Deploy Cert Manager
######################

data "template_file" "cert-manager-dns-issuer" {
  template = file("cert-manager-dns-issuer.yaml")
  vars = {
    dns_zones  = var.cluster_cloud_domain
    aws_region = var.aws_region
  }
}

resource "null_resource" "cert_manager_install" {
  triggers = {
    config_contents   = filemd5(var.config_path)
    k8s_yaml_contents = md5(data.template_file.cert-manager-dns-issuer.rendered)
  }
  provisioner "local-exec" {
    command = "kubectl apply --kubeconfig ${var.config_path} -f 'https://github.com/jetstack/cert-manager/releases/download/v0.13.0/cert-manager-no-webhook.yaml' && echo \"${data.template_file.cert-manager-dns-issuer.rendered}\" | kubectl apply -f - "
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
    when = "destroy"
    # destroy ingress object to remove created Load Balancer
    command = "kubectl delete --kubeconfig ${var.config_path} -f 'https://raw.githubusercontent.com/shalb/terraform-aws-minikube/master/addons/ingress.yaml'"
  }
}

######################
# Deploy ArgoCD
######################
locals {
  argocd_domain = "argocd.${var.cluster_cloud_domain}"
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

data "helm_repository" "argo" {
  name = "argo"
  url  = "https://argoproj.github.io/argo-helm"
}

# Deploy ArgoCD with Helm provider
resource "helm_release" "argo-cd" {
  name       = "argo-cd"
  repository = data.helm_repository.argo.metadata[0].name
  chart      = "argo-cd"
  version    = "2.0.0"
  namespace  = "argocd"

  values = [
    file("argocd-values.yaml")
  ]
  depends_on = [
    null_resource.kubeconfig_update,
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
