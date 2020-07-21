terraform {
  required_version = "~> 0.12.0"
  backend "s3" {
    region                      = "us-east-1"
    skip_credentials_validation = true
    skip_metadata_api_check     = true
  }
  required_providers {
    digitalocean = "~> 1.18.0"
    kubernetes   = "~> 1.11.0"
  }
}

data "terraform_remote_state" "vpc" {
  backend = "s3"
  config = {
    region                      = "us-east-1"
    endpoint                    = "${var.region}.digitaloceanspaces.com"
    skip_credentials_validation = true
    skip_metadata_api_check     = true
    bucket                      = var.cluster_name
    key                         = "states/terraform-vpc.state"
  }
  defaults = {
    vpc_id   = ""
    ip_range = ""
  }
}

provider "kubernetes" {
  load_config_file = false
  host             = digitalocean_kubernetes_cluster.k8s.endpoint
  token            = digitalocean_kubernetes_cluster.k8s.kube_config[0].token
  cluster_ca_certificate = base64decode(
    digitalocean_kubernetes_cluster.k8s.kube_config[0].cluster_ca_certificate
  )
}
