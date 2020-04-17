output "created_at" {
  description = "The date and time when the project was created, (ISO8601)"
  value       = digitalocean_project.project.id
}

output "id" {
  description = "The id of the project"
  value       = digitalocean_project.project.id
}

output "owner_id" {
  description = "The id of the project owner"
  value       = digitalocean_project.project.id
}

output "owner_uuid" {
  description = "The unique universal identifier of the project owner"
  value       = digitalocean_project.project.id
}

output "updated_at" {
  description = "The date and time when the project was last updated, (ISO8601)"
  value       = digitalocean_project.project.id
}
