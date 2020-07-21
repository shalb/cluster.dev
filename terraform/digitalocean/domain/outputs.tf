output "zone_id" {
  value = digitalocean_domain.sub.id
}
output "name_servers" {
  value = "ns1.digitalocean.com., ns2.digitalocean.com., ns3.digitalocean.com."
}
