terraform {
  required_version = "~> 0.12.0"
  backend "s3" {}

  required_providers {
    aws = "~> 2.23"
  }
}

provider "aws" {
  region = var.region
}
