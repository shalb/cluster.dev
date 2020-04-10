resource "digitalocean_spaces_bucket" "state" {
  name   = var.name
  region = var.region
  acl    = "private"

  lifecycle {
    prevent_destroy = true
  }
}
