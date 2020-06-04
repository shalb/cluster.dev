# Create EKS cluster
# TODO: support for managed node groups after they implement spot support https://github.com/aws/containers-roadmap/issues/583
module "eks" {
  source                               = "terraform-aws-modules/eks/aws"
  version                              = "12.0.0"
  cluster_name                         = var.cluster_name
  cluster_version                      = var.cluster_version
  subnets                              = var.workers_subnets_type == "private" ? data.terraform_remote_state.vpc.outputs.private_subnets : data.terraform_remote_state.vpc.outputs.public_subnets
  tags                                 = var.tags
  vpc_id                               = data.terraform_remote_state.vpc.outputs.vpc_id
  worker_groups_launch_template        = var.worker_groups_launch_template
  worker_additional_security_group_ids = [aws_security_group.worker_group_mgmt.id]
  cluster_endpoint_private_access      = "true"
  cluster_endpoint_public_access       = "true"
  enable_irsa                          = "true"
  # TODO: redesign AWS_PROFILE to be set
  #kubeconfig_aws_authenticator_env_variables = {
  #  AWS_PROFILE = "${var.awsprofile}"
  #}
}

# Add security group
resource "aws_security_group" "worker_group_mgmt" {
  name_prefix = "worker_group_mgmt"
  description = "SG to be applied to all *nix machines"
  vpc_id      = data.terraform_remote_state.vpc.outputs.vpc_id

  ingress {
    from_port = 22
    to_port   = 22
    protocol  = "tcp"

    cidr_blocks = [data.terraform_remote_state.vpc.outputs.vpc_cidr]
  }
}
