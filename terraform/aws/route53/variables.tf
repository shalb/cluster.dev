variable "region" {
  type        = string
  description = "The AWS region."
}

variable "cluster_fullname" {
  type        = string
  description = "Full name of the cluster"
}

variable "cluster_domain" {
  type        = string
  description = "Default domain for cluster records"
}

variable "zone_delegation" {
  type        = string
  default     = false
  description = "If true - a NS records in cluster_domain(cluster.dev) to be created by external scripts"
}
