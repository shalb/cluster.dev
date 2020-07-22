terraform {
  required_version = "~> 0.12.20"
  backend "s3" {}
  required_providers {
    helm       = "~> 1.0"
    kubernetes = "~> 1.11"
    null       = "~> 2.1"
    random     = "~> 2.2"
    template   = "~> 2.1"
  }
}
provider "aws" {
  region = var.region
}
provider "helm" {
  kubernetes {
    config_path = var.config_path
  }
}
provider "kubernetes" {
  config_path = var.config_path
}

## Get remote states to use in roles
data "terraform_remote_state" "k8s" {
  backend = "s3"
  config = {
    bucket = var.cluster_name
    key    = "states/terraform-k8s.state"
    region = var.region
  }
  defaults = {
    cluster_id              = ""
    cluster_oidc_issuer_url = ""
  }
}

data "terraform_remote_state" "dns" {
  backend = "s3"
  config = {
    bucket = var.cluster_name
    key    = "states/terraform-dns.state"
    region = var.region
  }
  defaults = {
    zone_id = ""
  }
}
