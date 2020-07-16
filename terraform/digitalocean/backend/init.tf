terraform {
  required_version = "~> 0.12.0"

  required_providers {
    digitalocean = "~> 1.18.0"
  }
}

provider "digitalocean" {
  spaces_endpoint = "https://ams3.digitaloceanspaces.com"
}
