variable "region" {
  type        = string
  description = "The DigitalOcean region."
}

variable "cluster_name" {
  type        = string
  description = "Name of the cluster."
}

variable "ip_range" {
  type        = string
  description = "The range of IP addresses for the VPC in CIDR notation"
  default     = "10.8.0.0/18"
}

variable "vpc_id" {
  type        = string
  description = "Vpc ID, or create or default"
  default     = "create"
}
