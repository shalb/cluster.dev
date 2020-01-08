provider "aws" {
  version    = ">= 2.23.0"
  region = var.region
}

resource "aws_default_subnet" "default" {
  availability_zone = "${var.region}a"  # TODO check if always default zone has A zone
  tags = {
    Name = "Default subnet for cluster.dev"
  }
}

data "template_file" "k8s_userdata" {
  template = "${file("k8s-userdata.tpl.sh")}"
  vars = {
    cluster_name = "${var.cluster_name}"
  }
}

module "minikube" {
  #source  = "scholzj/minikube/aws"
  #version = "1.10.0"  
  source =  "git::https://github.com/shalb/terraform-aws-minikube.git"
  cluster_name = var.cluster_name
  aws_instance_type = var.aws_instance_type
  aws_region = var.region
  aws_subnet_id = aws_default_subnet.default.id 
  hosted_zone = var.hosted_zone
  additional_userdata = data.template_file.k8s_userdata.rendered
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
