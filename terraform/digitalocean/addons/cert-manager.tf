######################
# Deploy Cert Manager
######################

locals {
  cert-manager-version = "v0.14.1"
  crd-path             = "https://github.com/jetstack/cert-manager/releases/download"
}

resource "null_resource" "cert_manager_crds" {
  provisioner "local-exec" {
    command = "kubectl apply --kubeconfig ${var.config_path} -n cert-manager --validate=false -f ${local.crd-path}/${local.cert-manager-version}/cert-manager.crds.yaml"
  }
  provisioner "local-exec" {
    when    = destroy
    command = "exit 0"
  }
  triggers = {
    content = sha1("${local.crd-path}/${local.cert-manager-version}/cert-manager.crds.yaml")
  }
}

resource "kubernetes_namespace" "cert-manager" {
  metadata {
    name = "cert-manager"
  }
}
resource "helm_release" "cert_manager" {
  name          = "cert-manager"
  chart         = "cert-manager"
  repository    = "https://charts.jetstack.io"
  namespace     = "cert-manager"
  version       = local.cert-manager-version
  recreate_pods = true

  values = [templatefile("${path.module}/templates/cert-manager-values.yaml", {
  })]
  depends_on = [
    null_resource.cert_manager_crds,
    null_resource.nginx_ingress_install,
  ]
}

# Add Production Issuer with DNS validation
# First need to create secret to store token to access DO API
# https://cert-manager.io/docs/configuration/acme/dns01/digitalocean/
resource "kubernetes_secret" "digitalocean-dns" {
  metadata {
    name      = "digitalocean-dns"
    namespace = "cert-manager"
  }
  data = {
    access-token = var.do_token
  }
}

data "template_file" "clusterissuers_production" {
  template = file("templates/cert-manager-dns-issuer.yaml")
  vars = {
  }
}

resource "null_resource" "cert_manager_issuers" {
  depends_on = [helm_release.cert_manager]

  provisioner "local-exec" {
    command = "kubectl apply --kubeconfig ${var.config_path} -n cert-manager -f -<<EOF\n${data.template_file.clusterissuers_production.rendered}\nEOF"
  }

  provisioner "local-exec" {
    when    = destroy
    command = "kubectl delete --kubeconfig ${var.config_path} -n cert-manager -f -<<EOF\n${data.template_file.clusterissuers_production.rendered}\nEOF"
  }

  triggers = {
    contents_production = sha1(data.template_file.clusterissuers_production.rendered)
  }
}
