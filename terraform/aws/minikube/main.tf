provider "aws" {
  version    = ">= 2.23.0"
  region = var.region
}
# TODO: replace all locals with vars
locals {
   hosted_zone = "shalb.net"
}

resource "aws_default_vpc" "default" {
  tags = {
    Name = "Default VPC"
  }
}

resource "aws_default_subnet" "default" {
  availability_zone = "${var.region}a"  # TODO check if always default zone has A zone
  vpc_id = aws_default_vpc.default.id

  tags = {
    Name = "Default subnet for cluster.dev"
  }
}

module "minikube" {
  source  = "scholzj/minikube/aws"
  version = "1.10.0"  
  cluster_name = var.cluster_name
  aws_instance_type = var.aws_instance_type
  aws_region = var.region
  aws_subnet_id = aws_default_subnet.default.id 
  hosted_zone = local.hosted_zone

  tags = {
    Application = "${var.cluster_name}"
    CreatedBy   = "cluster.dev"
  }

  addons = [
    "https://raw.githubusercontent.com/scholzj/terraform-aws-minikube/master/addons/storage-class.yaml",
    "https://raw.githubusercontent.com/scholzj/terraform-aws-minikube/master/addons/metrics-server.yaml",
    "https://raw.githubusercontent.com/scholzj/terraform-aws-minikube/master/addons/dashboard.yaml",
    "https://raw.githubusercontent.com/scholzj/terraform-aws-minikube/master/addons/external-dns.yaml",
  ]
}
