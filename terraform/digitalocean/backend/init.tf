terraform {
  required_version = "~> 0.12.0"

  required_providers {
    digitalocean = "~> 1.18.0"
  }
}

# Despite DO region should be set in us-east-1: https://github.com/aws/aws-sdk-go/issues/2232#issuecomment-434838388
provider digitalocean {
    region                      = "us-east-1"
}
