output "zone_id" {
  value = aws_route53_zone.sub.zone_id
}
output "name_servers" {
  value = aws_route53_zone.sub.name_servers
}
