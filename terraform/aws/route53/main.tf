provider "aws" {
  version    = ">= 2.23.0"
  region = var.region
}

resource "aws_route53_zone" "main" {
  name = "${var.cluster_domain}"
}

resource "aws_route53_zone" "sub" {
  name = "${var.cluster_fullname}.${var.cluster_domain}"
}

resource "aws_route53_record" "sub-ns" {
  zone_id = "${aws_route53_zone.main.zone_id}"
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
