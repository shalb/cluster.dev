# If DNS zone is provided by user.
# Use it to create subzone
data "digitalocean_domain" "provided" {
  count = tobool(var.zone_delegation) ? 0 : 1
  name  = "${var.cluster_domain}"
}

resource "digitalocean_domain" "sub" {
  name = "${var.cluster_name}.${var.cluster_domain}"
}

resource "digitalocean_record" "sub-ns1" {
  domain = tobool(var.zone_delegation) ? digitalocean_domain.sub.name : data.digitalocean_domain.provided.0.name
  name   = "${var.cluster_name}.${var.cluster_domain}"
  type   = "NS"
  ttl    = "300"
  value  = "ns1.digitalocean.com."
}

resource "digitalocean_record" "sub-ns2" {
  domain = tobool(var.zone_delegation) ? digitalocean_domain.sub.name : data.digitalocean_domain.provided.0.name
  name   = "${var.cluster_name}.${var.cluster_domain}"
  type   = "NS"
  ttl    = "300"
  value  = "ns2.digitalocean.com."
}

resource "digitalocean_record" "sub-ns3" {
  domain = tobool(var.zone_delegation) ? digitalocean_domain.sub.name : data.digitalocean_domain.provided.0.name
  name   = "${var.cluster_name}.${var.cluster_domain}"
  type   = "NS"
  ttl    = "300"
  value  = "ns3.digitalocean.com."
}

# Delegate created zone via lambda
resource "null_resource" "zone_delegation" {
  count = tobool(var.zone_delegation) ? 1 : 0
  # Update zone when DNS records are updated
  triggers = {
    domain_name = digitalocean_domain.sub.name
  }
  provisioner "local-exec" {
    command = <<EOF
        curl -H "Content-Type: application/json" -d '{"Action": "CREATE", "UserName": "${var.cluster_name}", "NameServers": "ns1.digitalocean.com.,ns2.digitalocean.com.,ns3.digitalocean.com.", "ZoneID": "${digitalocean_domain.sub.urn}", "DomainName": "${var.cluster_domain}", "Email": "${var.email}"}' ${var.dns_manager_url}
      EOF
  }
  provisioner "local-exec" {
    when    = destroy
    command = <<EOF
        curl -H "Content-Type: application/json" -d '{"Action": "DELETE", "UserName": "${var.cluster_name}", "NameServers": "ns1.digitalocean.com.,ns2.digitalocean.com.,ns3.digitalocean.com.", "ZoneID": "${digitalocean_domain.sub.urn}", "DomainName": "${var.cluster_domain}", "Email": "${var.email}"}' ${var.dns_manager_url}
      EOF
  }
}
