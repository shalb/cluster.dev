variable "description" {
  type        = string
  description = "A description for this project"
  default = ""
}

variable "resources" {
  type        = list
  description = "(Optional) List with all resources to be part of this project"
  default     = []
}

variable "environment" {
  type        = map
  description = "(optional) Edit values below for every environment name that you want"
  default = {
    "dev" = "Development"
    "stg" = "Staging"
    "prd" = "Production"
  }
}

variable "name" {
  type        = string
  description = "(Required) Name for your project at Digital Ocean"
}

variable "purpose" {
  type        = string
  description = "The purpose for your project. (for example: k8s or any), default - Web Application"
  default = "Web Application"
}
