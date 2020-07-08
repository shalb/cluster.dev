# INFO Read about spaces here https://www.digitalocean.com/docs/spaces/
terraform {
  required_version = ">= 0.12.0"

  backend "s3" {
    region                      = "us-east-1"
    access_key                  = ""
    secret_key                  = ""
    bucket                      = ""
    key                         = ""
    endpoint                    = ""
    skip_requesting_account_id  = true
    skip_credentials_validation = true
    skip_get_ec2_platforms      = true
    skip_metadata_api_check     = true
  }

  required_providers {
    digitalocean = "~> 1.16.0"
  }
}
