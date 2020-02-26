# Create a separate permanent backend for terraform state and configs: https://stackoverflow.com/a/48362341
# TODO Research encryption, ex: https://github.com/opendatacube/datacube-k8s-eks/blob/master/examples/quickstart/backend/terraform-backend.tf
provider "aws" {
  region = var.region
}

resource "aws_s3_bucket" "terraform_state" {
  bucket = var.s3_backend_bucket

  versioning {
    enabled = true
  }

  lifecycle {
    prevent_destroy = true
  }
}

resource "aws_dynamodb_table" "terraform_state_lock" {
  name           = "${var.s3_backend_bucket}-state"
  read_capacity  = 1
  write_capacity = 1
  hash_key       = "LockID"

  attribute {
    name = "LockID"
    type = "S"
  }
}
