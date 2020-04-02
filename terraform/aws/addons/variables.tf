variable "aws_region" {
  type        = string
  description = "AWS Region to apply for Addons configuration"
}

variable "cluster_cloud_domain" {
  type        = string
  description = "Route 53 zone used as a domain restrictions for cert-manager and external-dns"
  default     = ""
}

variable "config_path" {
  description = "path to a kubernetes config file"
  default     = "~/.kube/config"
}
