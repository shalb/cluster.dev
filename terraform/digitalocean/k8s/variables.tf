variable "cluster_name" {
  type        = string
  description = "(Required) Provide DigitalOcean cluster name"
}

variable "region" {
  type        = string
  description = "(Required) Provide DigitalOcean region"
}

variable "k8s_version" {
  type        = string
  description = "Provide DigitalOcean Kubernetes minor version (e.g. '1.15' or higher)"
  default     = "1.17"
}

variable "node_type" {
  description = "Digital Ocean Kubernetes default node pool type (e.g. `s-1vcpu-2gb` => 1vCPU, 2GB RAM)"
  type        = string
  default     = "s-1vcpu-2gb"
}

variable "min_node_count" {
  description = "Digital Ocean Kubernetes min nodes with autoscale feature (e.g. `1`)"
  type        = number
  default     = 1
}

variable "max_node_count" {
  description = "Digital Ocean Kubernetes max nodes with autoscale feature (e.g. `2`)"
  type        = number
  default     = 2
}

variable "write_kubeconfig" {
  description = "Whether to write a Kubectl config file containing the cluster configuration. Saved to `config_output_path`."
  type        = bool
  default     = true
}

variable "config_output_path" {
  description = "Where to save the Kubectl config file (if `write_kubeconfig = true`). Assumed to be a directory if the value ends with a forward slash `/`."
  type        = string
  default     = "./"
}
