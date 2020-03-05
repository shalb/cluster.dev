variable "region" {
  type        = string
  description = "The AWS region."
}

variable "cluster_name" {
  type        = string
  description = "Name of the cluster"
}

variable "vpc_cidr" {
  type        = string
  description = "Vpc CIDR"
  default     = "10.18.0.0/16"
}
