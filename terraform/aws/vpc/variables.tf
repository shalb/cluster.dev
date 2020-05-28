variable "region" {
  type        = string
  description = "The AWS region."
}

variable "availability_zones" {
  type        = list
  description = "The AWS Availability Zone(s) inside region."
}

variable "cluster_name" {
  type        = string
  description = "Name of the cluster."
}

variable "vpc_cidr" {
  type        = string
  description = "Vpc CIDR"
  default     = "10.8.0.0/18"
}

variable "vpc_id" {
  type        = string
  description = "Vpc ID, or create or default"
  default     = "default"
}
