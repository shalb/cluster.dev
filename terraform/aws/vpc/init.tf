terraform {
  backend "s3" {}

  required_providers {
    aws = "~> 2.64.0"
  }
}
provider "aws" {
  region = var.region
}
