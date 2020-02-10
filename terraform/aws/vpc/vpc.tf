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

  private_subnets = ["10.18.0.0/20", "10.18.16.0/20"]
//  private_subnet_tags = {
//    "kubernetes.io/role/internal-elb" = "1"
//  }

  public_subnets = ["10.18.128.0/20", "10.18.144.0/20"]
//  public_subnet_tags = {
////    "kubernetes.io/cluster/${var.awsprofile}-${var.environment}" = "owned"
//    "kubernetes.io/role/elb"                                     = "1"
//  }
  tags = {
    Terraform   = "true"
//    Environment = "${var.environment}"
  }
}


output "public_subnet_ids" {
  value = join(",", module.vpc.public_subnets)
}
