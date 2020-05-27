terraform {
  backend "s3" {}

  required_providers {
    aws = "~> 2.23"
  }
}
