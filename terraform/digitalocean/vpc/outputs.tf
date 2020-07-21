# If vpc_id =
# "create" - dispay vpc_id of created by module
# "default" - displey vpc_id of default vpc in the provided region
# "vpc_id"  - existing vpc - display just provided id
output "vpc_id" {
  value = var.vpc_id == "create" ? digitalocean_vpc.create[0].id : (var.vpc_id == "default" ? data.digitalocean_vpc.default[0].id : var.vpc_id)
}

output "ip_range" {
  value = var.vpc_id == "create" ? digitalocean_vpc.create[0].ip_range : (var.vpc_id == "default" ? data.digitalocean_vpc.default[0].ip_range : data.digitalocean_vpc.provided[0].ip_range)
}
