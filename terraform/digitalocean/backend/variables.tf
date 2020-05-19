variable "do_spaces_backend_bucket" {
  type        = string
  description = "(Required) Name for Spaces DO terraform backend"
}

variable "region" {
  type        = string
  description = "(Required) Region for Spaces DO terraform backend"
}

variable "digitalocean_token" {}
variable "spaces_access_id" {}
variable "spaces_secret_key" {}
