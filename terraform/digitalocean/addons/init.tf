terraform {
  required_version = "~> 0.12.20"
  backend "s3" {
    region                      = "us-east-1"
    skip_credentials_validation = true
    skip_metadata_api_check     = true
  }
  required_providers {
    helm       = "~> 1.0"
    kubernetes = "~> 1.11"
    null       = "~> 2.1"
    random     = "~> 2.2"
    template   = "~> 2.1"
  }
}

provider "helm" {
  kubernetes {
    config_path = var.config_path
  }
}

provider "kubernetes" {
  config_path = var.config_path
}

data "terraform_remote_state" "dns" {
  backend = "s3"
  config = {
    region                      = "us-east-1"
    endpoint                    = "${var.region}.digitaloceanspaces.com"
    skip_credentials_validation = true
    skip_metadata_api_check     = true
    bucket                      = var.cluster_name
    key                         = "states/terraform-dns.state"
  }
  defaults = {
    zone_id = ""
  }
}
