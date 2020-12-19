terraform {
  required_version = "~> 0.13.00"
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
