provider "aws" {
  region = var.region
}

resource "aws_route53_zone" "main" {
  name = var.cluster_domain
}

resource "aws_route53_zone" "sub" {
  name = "${var.cluster_fullname}.${var.cluster_domain}"
  force_destroy = true
}

resource "aws_route53_record" "sub-ns" {
  allow_overwrite = true
  zone_id = aws_route53_zone.main.zone_id
  name    = "${var.cluster_fullname}.${var.cluster_domain}"
  type    = "NS"
  ttl     = "300"

  records = [
    "${aws_route53_zone.sub.name_servers.0}",
    "${aws_route53_zone.sub.name_servers.1}",
    "${aws_route53_zone.sub.name_servers.2}",
    "${aws_route53_zone.sub.name_servers.3}",
  ]
}

# Delegate created zone via lambda
resource "null_resource" "zone_delegation" {
  count    = "${var.zone_delegation == "true" ? 1 : 0}"
# Update zone when DNS records are updated
  triggers = {
    zone_id = "${aws_route53_zone.sub.zone_id}"
    dns_record = "${aws_route53_zone.sub.name_servers.0}"
  }
  provisioner "local-exec" {
      command = <<EOF
        curl -H "Content-Type: application/json" -d '{"Action": "CREATE", "UserName": "${var.cluster_fullname}", "NameServers": "${aws_route53_zone.sub.name_servers.0}.,${aws_route53_zone.sub.name_servers.1}.,${aws_route53_zone.sub.name_servers.2}.,${aws_route53_zone.sub.name_servers.3}", "ZoneID": "${aws_route53_zone.sub.zone_id}", "DomainName": "${var.cluster_domain}", "Email": "voa@shalb.com"}' https://usgrtk5fqj.execute-api.eu-central-1.amazonaws.com/prod
      EOF
  }
  provisioner "local-exec" {
    when = destroy
      command = <<EOF
        curl -H "Content-Type: application/json" -d '{"Action": "DELETE", "UserName": "${var.cluster_fullname}", "NameServers": "${aws_route53_zone.sub.name_servers.0}.,${aws_route53_zone.sub.name_servers.1}.,${aws_route53_zone.sub.name_servers.2}.,${aws_route53_zone.sub.name_servers.3}", "ZoneID": "${aws_route53_zone.sub.zone_id}", "DomainName": "${var.cluster_domain}", "Email": "voa@shalb.com"}' https://usgrtk5fqj.execute-api.eu-central-1.amazonaws.com/prod
      EOF
  }
}
