terraform {
  backend "s3" {}
  required_providers {
    helm = "~> 1.0"
    kubernetes = "~> 1.11"
    null = "~> 2.1"
    random = "~> 2.2"
    template = "~> 2.1"
  }
}
provider "aws" {
  region = var.region
}
provider "helm" {
}

## Get remote states to use in roles
data "terraform_remote_state" "eks" {
  backend = "s3"
  config = {
    bucket = var.cluster_name
    key    = "states/terraform-eks.state"
    region = var.region
  }
}

data "terraform_remote_state" "dns" {
  backend = "s3"
  config = {
    bucket = var.cluster_name
    key    = "states/terraform-dns.state"
    region = var.region
  }
}
