terraform {
  backend "local" {}
  required_providers {
    digitalocean = "~> 1.16.0"
  }
}
