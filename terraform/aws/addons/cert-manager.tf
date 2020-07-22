######################
# Deploy Cert Manager
######################

locals {
  cert-manager-version = "v0.14.1"
  crd-path             = "https://github.com/jetstack/cert-manager/releases/download"
}

resource "null_resource" "cert_manager_crds" {
  provisioner "local-exec" {
    command = "kubectl apply --kubeconfig ${var.config_path} -n cert-manager --validate=false -f ${local.crd-path}/${local.cert-manager-version}/cert-manager.crds.yaml"
  }
  provisioner "local-exec" {
    when    = destroy
    command = "exit 0"
  }
  triggers = {
    content = sha1("${local.crd-path}/${local.cert-manager-version}/cert-manager.crds.yaml")
  }
}

resource "kubernetes_namespace" "cert-manager" {
  metadata {
    name = "cert-manager"
  }
}
resource "helm_release" "cert_manager" {
  name          = "cert-manager"
  chart         = "cert-manager"
  repository    = "https://charts.jetstack.io"
  namespace     = "cert-manager"
  version       = local.cert-manager-version
  recreate_pods = true

  values = [templatefile("${path.module}/templates/cert-manager-values.yaml", {
    eks_service_account = module.iam_assumable_role_cert_manager.this_iam_role_arn
  })]
  depends_on = [
    null_resource.cert_manager_crds,
    kubernetes_namespace.cert-manager,
  ]
}

# Add IRSA role
module "iam_assumable_role_cert_manager" {
  source                        = "terraform-aws-modules/iam/aws//modules/iam-assumable-role-with-oidc"
  version                       = "~> v2.6.0"
  create_role                   = var.eks ? true : false
  role_name                     = "eks-cert-manager-${data.terraform_remote_state.k8s.outputs.cluster_id}-${random_id.random_role_id.hex}"
  provider_url                  = replace(data.terraform_remote_state.k8s.outputs.cluster_oidc_issuer_url, "https://", "")
  role_policy_arns              = [var.eks && length(aws_iam_policy.cert_manager) >= 1 ? aws_iam_policy.cert_manager.0.arn : ""]
  oidc_fully_qualified_subjects = ["system:serviceaccount:cert-manager:cert-manager"]
}

resource "aws_iam_policy" "cert_manager" {
  count = var.eks ? 1 : 0

  name   = "AllowCertManagerUpdates-${data.terraform_remote_state.k8s.outputs.cluster_id}-${random_id.random_role_id.hex}"
  policy = <<-EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": "route53:GetChange",
      "Resource": "arn:aws:route53:::change/*"
    },
    {
      "Effect": "Allow",
      "Action": [
        "route53:ChangeResourceRecordSets",
        "route53:ListResourceRecordSets"
      ],
      "Resource": "arn:aws:route53:::hostedzone/*"
    },
    {
      "Effect": "Allow",
      "Action": "route53:ListHostedZonesByName",
      "Resource": "*"
    }
  ]
}
  EOF
}

# Add Production Issuer with DNS validation
data "template_file" "clusterissuers_production" {
  template = file("templates/cert-manager-dns-issuer.yaml")
  vars = {
    dns_zones = var.cluster_cloud_domain
    region    = var.region
    # Do not need role with IRSA/EKS is present
    role = ""
  }
}


resource "null_resource" "cert_manager_issuers" {
  depends_on = [helm_release.cert_manager]

  provisioner "local-exec" {
    command = "kubectl apply --kubeconfig ${var.config_path} -n cert-manager -f -<<EOF\n${data.template_file.clusterissuers_production.rendered}\nEOF"
  }

  provisioner "local-exec" {
    when    = destroy
    command = "kubectl delete --kubeconfig ${var.config_path} -n cert-manager -f -<<EOF\n${data.template_file.clusterissuers_production.rendered}\nEOF"
  }

  triggers = {
    contents_production = sha1(data.template_file.clusterissuers_production.rendered)
  }
}
