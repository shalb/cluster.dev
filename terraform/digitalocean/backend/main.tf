resource "digitalocean_spaces_bucket" "terraform_state" {
  name   = var.do_spaces_backend_bucket
  region = var.region
  acl    = "private"

  token             = var.do_token
  spaces_access_id  = var.access_id
  spaces_secret_key = var.secret_key

  lifecycle {
    prevent_destroy = true
  }
}
