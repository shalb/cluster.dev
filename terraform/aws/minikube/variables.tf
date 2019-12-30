variable "region" {
  type        = string
  description = "The AWS region."
}

variable "cluster_name" {
  type        = string
  description = "Name of the cluster"
}

variable "aws_instance_type" {
  type = string
}
