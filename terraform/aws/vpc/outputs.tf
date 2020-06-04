# If vpc_id =
# "create" - dispay vpc_id of created by module
# "default" - displey vpc_id of default vpc in the provided region
# "vpc_id"  - existing vpc - display just provided id
output "vpc_id" {
  value = var.vpc_id == "create" ? module.vpc.vpc_id : (var.vpc_id == "default" ? aws_default_vpc.default[0].id : var.vpc_id)
}

output "private_subnets" {
  value = var.vpc_id == "create" ? module.vpc.private_subnets : (var.vpc_id == "default" ? [aws_default_subnet.default_az0[0].id, aws_default_subnet.default_az1[0].id] : data.aws_subnet_ids.vpc_subnets_provided_private[0].ids)
}

output "public_subnets" {
  value = var.vpc_id == "create" ? module.vpc.public_subnets : (var.vpc_id == "default" ? [aws_default_subnet.default_az0[0].id, aws_default_subnet.default_az1[0].id] : data.aws_subnet_ids.vpc_subnets_provided_public[0].ids)
}

output "vpc_cidr" {
  value = var.vpc_id == "create" ? module.vpc.vpc_cidr_block : (var.vpc_id == "default" ? aws_default_vpc.default[0].cidr_block : data.aws_vpc.provided[0].cidr_block)
}
