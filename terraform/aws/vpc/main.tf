module "vpc" {
  source  = "terraform-aws-modules/vpc/aws"
  version = "2.15.0"

  name               = "${var.cluster_name}-vpc"
  cidr               = "10.18.0.0/16"
  azs                = ["${var.region}a"]
  enable_nat_gateway = true
  enable_vpn_gateway = true

  enable_dns_hostnames = true
  enable_dns_support   = true

  public_subnets = ["10.18.128.0/20"]
  tags = {
    Terraform = "true"
  }
}

output "vpc_id" {
  value = module.vpc.vpc_id
}
