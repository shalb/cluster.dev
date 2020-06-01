terraform {
  backend "s3" {}

  required_providers {
    aws = "~> 2.64.0"
    null = "~> 2.1"
  }
}
provider "aws" {
  region = var.region
}
