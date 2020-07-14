resource "digitalocean_spaces_bucket" "terraform_state" {
  name   = var.do_spaces_backend_bucket
  region = var.region
  acl    = "private"

  lifecycle {
    prevent_destroy = true
  }
}
