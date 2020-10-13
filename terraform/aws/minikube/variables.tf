variable "region" {
  type        = string
  description = "The AWS region."
}

variable "cluster_name" {
  type        = string
  description = "Name of the cluster"
}

variable "aws_instance_type" {
  type        = string
  description = "Instance size"
}

variable "hosted_zone" {
  type        = string
  description = "DNS zone to use in cluster"
}
