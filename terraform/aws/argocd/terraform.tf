terraform {
  backend "s3" {}

  required_providers {
    helm = "~> 1.0"
  }
}
