variable "region" {
  type        = string
  description = "AWS Region to apply for Addons configuration"
}

variable "cluster_cloud_domain" {
  type        = string
  description = "Route 53 zone used as a domain restrictions for cert-manager and external-dns"
  default     = ""
}

variable "cluster_name" {
  type        = string
  description = "Full cluster name including user/organization"
  default     = ""
}

variable "config_path" {
  type        = string
  description = "path to a kubernetes config file"
  default     = "~/.kube/config"
}

variable "eks" {
  type        = bool
  description = "Define if addons would be deployed to EKS cluster"
  default     = false
}
