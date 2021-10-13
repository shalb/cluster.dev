variable "data" {
  type = string
}
variable "bucket_name" {
  type = string
}
resource "aws_s3_bucket_object" "cdev_object" {
  key     = "cdevs3data"
  bucket  = var.bucket_name
  content = var.data

  force_destroy = true
}

output "cantent" {
  value = var.data
}

output "file_s3_url" {
  value = "s3://${var.bucket_name}/cdevs3data"
}
