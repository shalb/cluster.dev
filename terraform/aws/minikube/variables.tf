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

variable "vpc_id" {
  type        = string
  description = "VPC ID (In case it differs from default)"
  default     = ""
}

