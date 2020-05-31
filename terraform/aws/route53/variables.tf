variable "region" {
  type        = string
  description = "The AWS region."
}

variable "cluster_name" {
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

variable "dns_manager_url" {
  type        = string
  default     = "https://usgrtk5fqj.execute-api.eu-central-1.amazonaws.com/prod"
  description = "Endpoint to create a default zone in cluster.dev domain"
}

variable "email" {
  type        = string
  default     = "domain-request@cluster.dev"
  description = "email of user which requests a default zone"
}
