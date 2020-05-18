locals {}

provider "aws" {
  region = var.region
}

module "vpc" {
  source  = "terraform-aws-modules/vpc/aws"
  version = "2.15.0"

  name               = "${var.cluster_name}-vpc"
  cidr               = var.vpc_cidr
  azs                = var.availability_zones
  enable_nat_gateway = true
  enable_vpn_gateway = true

  enable_dns_hostnames = true
  enable_dns_support   = true
  public_subnets       = [for k, v in var.availability_zones : cidrsubnet(var.vpc_cidr, 4, k)]
  public_subnet_tags = {
    "kubernetes.io/cluster/${var.cluster_name}" = "owned"
    "kubernetes.io/role/elb"                    = "1"
    "cluster.dev/subnet_type" = "public"
  }
  private_subnets = [for k, v in var.availability_zones : cidrsubnet(var.vpc_cidr, 4, k + 4)]
  private_subnet_tags = {
    "kubernetes.io/role/internal-elb" = "1"
    "cluster.dev/subnet_type" = "private"
  }
# database_subnets could be created in case of AZ's > 1
  database_subnets = length(var.availability_zones) > 1 ? [for k, v in var.availability_zones : cidrsubnet(var.vpc_cidr, 4, k + 8)] : []
  tags = {
    Terraform                  = "true"
    "cluster.dev/subnet_type" = "database"
  }
}

output "vpc_id" {
  value = module.vpc.vpc_id
}
