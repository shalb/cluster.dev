data "template_file" "auth_info" {
  template = file("templates/auth.tmpl")
  vars = {
    argocd_url      = "https://${local.argocd_domain}"
    argocd_login    = "admin"
    argocd_password = random_password.argocd_pass.result
  }
}

resource "digitalocean_spaces_bucket_object" "auth_file" {
  region  = var.region
  bucket  = var.cluster_name
  key     = "addons/auth.yaml"
  content = data.template_file.auth_info.rendered
}
