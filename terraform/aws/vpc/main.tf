# If vpc_id = "create" use module to create new VPC
module "vpc" {
  create_vpc = (var.vpc_id == "create" ? "true" : "false")
  source     = "terraform-aws-modules/vpc/aws"
  version    = "2.15.0"

  name               = "${var.cluster_name}-vpc"
  cidr               = var.vpc_cidr
  azs                = var.availability_zones
  enable_nat_gateway = true
  enable_vpn_gateway = false

  enable_dns_hostnames = true
  enable_dns_support   = true
  # in case when vpc cidr block provided by user, we split it to blocks for each subnet
  # which is relative to number of availability zones. See details on cidrsubnet in terraform docs
  # https://www.terraform.io/docs/configuration/functions/cidrsubnet.html
  public_subnets = [for subnet_num, az in var.availability_zones : cidrsubnet(var.vpc_cidr, 4, subnet_num)]
  public_subnet_tags = {
    "kubernetes.io/cluster/${var.cluster_name}" = "owned"
    "kubernetes.io/role/elb"                    = "1"
    "cluster.dev/subnet_type"                   = "public"
  }
  private_subnets = [for subnet_num, az in var.availability_zones : cidrsubnet(var.vpc_cidr, 4, subnet_num + 4)]
  private_subnet_tags = {
    "kubernetes.io/role/internal-elb" = "1"
    "cluster.dev/subnet_type"         = "private"
  }

  # TODO: add IAM policy to create DB Subnet
  # database_subnets could be created in case of AZ's > 1
  #  database_subnets = length(var.availability_zones) > 1 ? [for subnet_num, az in var.availability_zones : cidrsubnet(var.vpc_cidr, 4, subnet_num + 8)] : []
  #  tags = {
  #    Terraform                  = "true"
  #    "cluster.dev/subnet_type" = "database"
  #  }
}

# If vpc_id is provided ex('vpc-2b2212c') - use data objects to get outputs
data "aws_vpc" "provided" {
  id    = var.vpc_id
  count = var.vpc_id != "default" && var.vpc_id != "create" ? 1 : 0
}

# TODO: document that subnets of existing vpc should be tagged with next tag
data "aws_subnet_ids" "vpc_subnets_provided_private" {
  count  = var.vpc_id != "default" && var.vpc_id != "create" ? 1 : 0
  vpc_id = var.vpc_id
  tags = {
    "cluster.dev/subnet_type" = "private"
  }
}

data "aws_subnet_ids" "vpc_subnets_provided_public" {
  count  = var.vpc_id != "default" && var.vpc_id != "create" ? 1 : 0
  vpc_id = var.vpc_id
  tags = {
    "cluster.dev/subnet_type" = "public"
  }
}

# If vpc_id is not provided - use default VPC's as a resource
resource "aws_default_vpc" "default" {
  count = var.vpc_id == "default" ? 1 : 0
  tags = {
    Name = "Default VPC"
  }
}

resource "aws_default_subnet" "default_az0" {
  count             = var.vpc_id == "default" ? 1 : 0
  availability_zone = length(var.availability_zones) > 0 ? var.availability_zones[0] : "${var.region}a"
  tags = {
    Name                      = "Default subnet for cluster.dev in AZ1"
    "cluster.dev/subnet_type" = "default"
  }
}

resource "aws_default_subnet" "default_az1" {
  count             = var.vpc_id == "default" ? 1 : 0
  availability_zone = length(var.availability_zones) > 1 ? var.availability_zones[1] : "${var.region}b"
  tags = {
    Name                      = "Default subnet for cluster.dev in AZ2"
    "cluster.dev/subnet_type" = "default"
  }
}
