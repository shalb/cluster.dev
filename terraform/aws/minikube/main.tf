data "template_file" "k8s_userdata" {
  template = file("k8s-userdata.tpl.sh")
  vars = {
    cluster_name = var.cluster_name
    private_key  = tls_private_key.bastion_key.private_key_pem
  }
}

module "minikube" {
  source              = "git::https://github.com/shalb/terraform-aws-minikube.git?ref=v0.1.0"
  cluster_name        = var.cluster_name
  aws_instance_type   = var.aws_instance_type
  aws_region          = var.region
  aws_subnet_id       = data.terraform_remote_state.vpc.outputs.public_subnets[0]
  hosted_zone         = var.hosted_zone
  additional_userdata = data.template_file.k8s_userdata.rendered
  ssh_public_key      = tls_private_key.bastion_key.public_key_openssh # generated in bastion.tf
  tags = {
    Application = var.cluster_name
    CreatedBy   = "cluster.dev"
  }

  addons = [
    "https://raw.githubusercontent.com/shalb/terraform-aws-minikube/master/addons/storage-class.yaml",
    "https://raw.githubusercontent.com/shalb/terraform-aws-minikube/master/addons/metrics-server.yaml",
    "https://raw.githubusercontent.com/shalb/terraform-aws-minikube/master/addons/dashboard.yaml"
  ]

}
