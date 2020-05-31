# If DNS zone is provided by user.
# Use it to create subzone
data "aws_route53_zone" "existing" {
  count = tobool(var.zone_delegation) ? 0 : 1
  name  = "${var.cluster_domain}."
}

resource "aws_route53_zone" "sub" {
  name          = "${var.cluster_name}.${var.cluster_domain}"
  force_destroy = true
}

resource "aws_route53_record" "sub-ns" {
  allow_overwrite = true
  zone_id         = tobool(var.zone_delegation) ? aws_route53_zone.sub.zone_id : data.aws_route53_zone.existing.0.zone_id
  name            = "${var.cluster_name}.${var.cluster_domain}"
  type            = "NS"
  ttl             = "300"

  records = aws_route53_zone.sub.name_servers
}

# Delegate created zone via lambda
resource "null_resource" "zone_delegation" {
  count = tobool(var.zone_delegation) ? 1 : 0
  # Update zone when DNS records are updated
  triggers = {
    zone_id    = aws_route53_zone.sub.zone_id
    dns_record = aws_route53_zone.sub.name_servers.0
  }
  provisioner "local-exec" {
    command = <<EOF
        curl -H "Content-Type: application/json" -d '{"Action": "CREATE", "UserName": "${var.cluster_name}", "NameServers": "${aws_route53_zone.sub.name_servers.0}.,${aws_route53_zone.sub.name_servers.1}.,${aws_route53_zone.sub.name_servers.2}.,${aws_route53_zone.sub.name_servers.3}", "ZoneID": "${aws_route53_zone.sub.zone_id}", "DomainName": "${var.cluster_domain}", "Email": "${var.email}"}' ${var.dns_manager_url}
      EOF
  }
  provisioner "local-exec" {
    when    = destroy
    command = <<EOF
        curl -H "Content-Type: application/json" -d '{"Action": "DELETE", "UserName": "${var.cluster_name}", "NameServers": "${aws_route53_zone.sub.name_servers.0}.,${aws_route53_zone.sub.name_servers.1}.,${aws_route53_zone.sub.name_servers.2}.,${aws_route53_zone.sub.name_servers.3}", "ZoneID": "${aws_route53_zone.sub.zone_id}", "DomainName": "${var.cluster_domain}", "Email": "${var.email}"}' ${var.dns_manager_url}
      EOF
  }
}
