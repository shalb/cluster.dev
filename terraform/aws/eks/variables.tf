variable "region" {
  type        = string
  description = "The AWS region."
}
variable "availability_zones" {
  type        = list
  description = "The AWS Availability Zone(s) inside region."
}
variable "cluster_name" {
  description = "Name of the EKS cluster. Also used as a prefix in names of related resources."
  type        = string
}
variable "workers_subnets_type" {
  description = "Type of subnets to use on worker nodes: public or private"
  type        = string
  default     = "private"
}
variable "cluster_version" {
  description = "Kubernetes version to use for the EKS cluster."
  type        = string
  default     = "1.16"
}
variable "vpc_id" {
  description = "VPC where the cluster and workers will be deployed."
  type        = string
}
variable "worker_additional_security_group_ids" {
  description = "A list of additional security group ids to attach to worker instances"
  type        = list(string)
  default     = []
}
variable "tags" {
  description = "A map of tags to add to all resources."
  type        = map(string)
  default     = {}
}
variable "worker_groups_launch_template" {
  description = "A list of maps defining worker group configurations to be defined using AWS Launch Templates. See workers_group_defaults for valid keys."
  type        = any
  default     = []
}
