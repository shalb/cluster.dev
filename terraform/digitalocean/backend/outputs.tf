output "spaces_bucket_name" {
  description = "Digital Ocean Spaces bucket name"
  value       = digitalocean_spaces_bucket.terraform_state.name
}
