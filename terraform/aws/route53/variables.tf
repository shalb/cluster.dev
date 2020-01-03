variable "region" {
  type        = string
  description = "The AWS region."
}

variable "cluster_fullname" {
  type        = string
  description = "Full name of the cluster"
}

variable "cluster_domain" {
  type = string
  description = "Default domain for cluster records"
}
