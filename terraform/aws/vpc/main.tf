locals {}

provider "aws" {
  version = ">= 2.23.0"
  region  = var.region
}

module "vpc" {
  source  = "terraform-aws-modules/vpc/aws"
  version = "2.15.0"

  name               = "${var.cluster_name}-vpc"
  cidr               = var.vpc_cidr
  azs                = ["${var.region}a"]
  enable_nat_gateway = true
  enable_vpn_gateway = true

  enable_dns_hostnames = true
  enable_dns_support   = true
  public_subnets  = [cidrsubnet(var.vpc_cidr, 4, 0)]
  tags = {
    Terraform = "true"
  }
}

output "vpc_id" {
  value = module.vpc.vpc_id
}
