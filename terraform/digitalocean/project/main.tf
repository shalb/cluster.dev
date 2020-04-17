resource "digitalocean_project" "project" {
  name        = var.name
  description = var.description
  purpose     = var.purpose
  environment = var.environment
  resources   = var.resources
}
