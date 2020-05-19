variable "do_spaces_backend_bucket" {
  type        = string
  description = "(Required) Name for Spaces DO terraform backend"
}

variable "region" {
  type        = string
  description = "(Required) Region for Spaces DO terraform backend"
}

variable "do_token" {}
variable "do_access_key" {}
variable "do_secret_key" {}
