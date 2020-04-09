variable "name" {
  type        = string
  description = "(Required) Name for Spaces DO terraform backend"
}

variable "region" {
  type = string
  description = "(Required) Region for Spaces DO terraform backend"
}

variable "do_token" {
  type        = string
  description = "Digital Ocean personal access token."
}

variable "spaces_access_id" {
  type        = string
  description = "Digital Ocean Spaces access id."
}

variable "spaces_secret_key" {
  type        = string
  description = "Digital Ocean Spaces secret key."
}
