terraform {
  required_version = "~> 0.12.0"
  backend "s3" {
    region                      = "us-east-1"
    skip_credentials_validation = true
    skip_metadata_api_check     = true
  }
  required_providers {
    digitalocean = "~> 1.18.0"
  }
}
