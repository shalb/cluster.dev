resource "digitalocean_spaces_bucket" "terraform_state" {
  name   = var.do_spaces_backend_bucket
  region = "us-east-1"
  acl    = "private"

  lifecycle {
    prevent_destroy = true
  }
}
