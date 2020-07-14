variable "do_spaces_backend_bucket" {
  type        = string
  description = "(Required) Name for Spaces DO terraform backend"
}

variable "region" {
  type        = string
  description = "(Required) Region for Spaces DO terraform backend"
}

variable "do_token" {
  type        = string
}

variable "access_id" {
  type        = string
}

variable "secret_key" {
  type        = string
}
