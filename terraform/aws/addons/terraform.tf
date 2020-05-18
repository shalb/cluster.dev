terraform {
  required_version = ">= 0.12.0"

  backend "s3" {}

  required_providers {
    helm = "~> 1.0"
  }
}
