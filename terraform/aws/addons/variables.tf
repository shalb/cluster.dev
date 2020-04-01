variable "aws_region" {
  type = string
  description = "AWS Region to apply for Addons configuration"
}

variable "cluster_cloud_zone" {
  type = string
  description = "Route 53 zone used as a domain restrictions for cert-manager and external-dns"
  default = ""
}

