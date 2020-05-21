provider "aws" {
  region = var.region
}

# If VPC is provided - use data objects
data "aws_vpc" "provided" {
  id    = var.vpc_id
  count = var.vpc_id != "" ? 1 : 0
}

data "aws_subnet_ids" "vpc_subnets" {
  count  = var.vpc_id != "" ? 1 : 0
  vpc_id = var.vpc_id
  tags = {
    "cluster.dev/subnet_type" = "private"
  }
}

data "aws_subnet" "private" {
  for_each = data.aws_subnet_ids.vpc_subnets[0].ids
  id       = each.value
}

# If VPC not provided - use default one
resource "aws_default_vpc" "default" {
  count = var.vpc_id != "" ? 0 : 1
  tags = {
    Name = "Default VPC"
  }
}

resource "aws_default_subnet" "default" {
  count             = var.vpc_id != "" ? 0 : 1
  availability_zone = "${var.region}a"
  tags = {
    Name                      = "Default subnet for cluster.dev"
    "cluster.dev/subnet_type" = "default"
  }
}

# Create EKS cluster
# TODO: support for managed node groups after they implement spot support https://github.com/aws/containers-roadmap/issues/583
module "eks" {
  source                               = "terraform-aws-modules/eks/aws"
  version                              = "12.0.0"
  cluster_name                         = var.cluster_name
  cluster_version                      = var.cluster_version
  subnets                              = var.vpc_id != "" ? [for s in data.aws_subnet.private : s.id] : aws_default_subnet.default[0].id
  tags                                 = var.tags
  vpc_id                               = var.vpc_id != "" ? var.vpc_id : aws_default_vpc.default[0].id
  worker_groups_launch_template        = var.worker_groups_launch_template
  worker_additional_security_group_ids = [aws_security_group.worker_group_mgmt.id]
  cluster_endpoint_private_access      = "false"
  cluster_endpoint_public_access       = "true"
  # TODO: redesign AWS_PROFILE to be set
  #kubeconfig_aws_authenticator_env_variables = {
  #  AWS_PROFILE = "${var.awsprofile}"
  #}
}

# Add security group
resource "aws_security_group" "worker_group_mgmt" {
  name_prefix = "worker_group_mgmt"
  description = "SG to be applied to all *nix machines"
  vpc_id      = var.vpc_id

  ingress {
    from_port = 22
    to_port   = 22
    protocol  = "tcp"

    cidr_blocks = var.vpc_id != "" ? [data.aws_vpc.provided[0].cidr_block] : [aws_default_vpc.default[0].cidr_block]
  }
}

