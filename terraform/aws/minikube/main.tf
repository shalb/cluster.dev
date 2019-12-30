locals {
   aws_subnet_id = "subnet-304b7f7a"
   hosted_zone = "shalb.net"
}

provider "aws" {
  version    = ">= 2.23.0"
  region = var.region
}

module "minikube" {
  source  = "scholzj/minikube/aws"
  version = "1.10.0"  
  cluster_name = var.cluster_name
  aws_instance_type = var.aws_instance_type
  aws_region = var.region
  aws_subnet_id = local.aws_subnet_id
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
