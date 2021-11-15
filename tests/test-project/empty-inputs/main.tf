
resource "aws_s3_bucket" "testemptyinputs" {
  bucket        = "testemptyinputs"
  force_destroy = true
  tags = {
    Name = "Cdev auto tests bucket"
  }
}

output "id" {
  value = aws_s3_bucket.testemptyinputs.id
}
