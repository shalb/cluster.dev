terraform {
  required_version = "~> 0.12.20"
  backend "s3" {}

  required_providers {
    random     = "~> 2.2.1"
    kubernetes = "~> 1.11.3"
    aws        = "~> 2.64.0"
    local      = "~> 1.4.0"
    null       = "~> 2.1.2"
  }
}

provider "aws" {
  region = var.region
}

data "terraform_remote_state" "vpc" {
  backend = "s3"
  config = {
    bucket = var.cluster_name
    key    = "states/terraform-vpc.state"
    region = var.region
  }
  defaults = {
    private_subnets = []
    public_subnets  = []
    vpc_id          = ""
    vpc_cidr        = ""
  }
}
