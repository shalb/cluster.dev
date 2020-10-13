data "template_file" "auth_info" {
  template = file("templates/auth.tmpl")
  vars = {
    argocd_url      = "https://${local.argocd_domain}"
    argocd_login    = "admin"
    argocd_password = random_password.argocd_pass.result
  }
}

resource "local_file" "auth_file" {
  content  = data.template_file.auth_info.rendered
  filename = "${path.module}/auth.yaml"
}

resource "aws_s3_bucket_object" "auth_upload" {
  bucket     = var.cluster_name
  key        = "addons/auth.yaml"
  source     = "${path.module}/auth.yaml"
  depends_on = [local_file.auth_file]
}
