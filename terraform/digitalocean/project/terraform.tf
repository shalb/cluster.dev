# INFO Read about spaces here https://www.digitalocean.com/docs/spaces/
terraform {
  backend "s3" {
    region = "us-west-1"
    skip_requesting_account_id  = true
    skip_credentials_validation = true
    skip_get_ec2_platforms      = true
    skip_metadata_api_check     = true
  }

  required_providers {
    digitalocean = "~> 1.16.0"
  }
}
