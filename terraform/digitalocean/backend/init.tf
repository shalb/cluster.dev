terraform {
  required_version = "~> 0.12.0"

  required_providers {
    digitalocean = "~> 1.18.0"
  }
}

provider digitalocean {

  token             = var.do_token
  spaces_access_id  = var.access_id
  spaces_secret_key = var.secret_key

}
