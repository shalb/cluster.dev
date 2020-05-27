#####################
# Deploy ExternalDNS
#####################

# Check for kubeconfig and update resources when cluster gets re-created.
resource "null_resource" "kubeconfig_update" {
  triggers = {
    policy_sha1 = "${sha1(file(var.config_path))}"
  }
}

# Create Namespace for ExternalDNS
resource "kubernetes_namespace" "external-dns" {
  metadata {
    name = "external-dns"
  }
}

data "helm_repository" "bitnami" {
  name = "bitnami"
  url  = "https://charts.bitnami.com/bitnami"
}

resource "helm_release" "external-dns" {
  name       = "external-dns"
  repository = data.helm_repository.bitnami.metadata[0].name
  chart      = "external-dns"
  version    = "2.20.10"
  namespace  = "external-dns"

  #values = [
  #  file("external-dns-values.yaml")
  #]
  depends_on = [
    null_resource.kubeconfig_update,
    kubernetes_namespace.external-dns,
  ]
  set {
    name  = "aws.region"
    value = var.region
  }
  set {
    name  = "rbac.serviceAccountAnnotations.eks\\.amazonaws\\.com/role-arn"
    value = module.iam_assumable_role_external_dns.this_iam_role_arn
  }
}

## Get remote states to use in roles
data "terraform_remote_state" "eks" {
  backend = "s3"
  config = {
    bucket = var.cluster_fullname
    key    = "dev-eks/terraform-eks.state"
    region = var.region
  }
}

data "terraform_remote_state" "dns" {
  backend = "s3"
  config = {
    bucket = var.cluster_fullname
    key    = "dev-eks/terraform-dns.state"
    region = var.region
  }
}

module "iam_assumable_role_external_dns" {
  source                        = "terraform-aws-modules/iam/aws//modules/iam-assumable-role-with-oidc"
  version                       = "~> v2.6.0"
  create_role                   = true
  role_name                     = "eks-external-dns-${data.terraform_remote_state.eks.outputs.cluster_id}"
  provider_url                  = replace(data.terraform_remote_state.eks.outputs.cluster_oidc_issuer_url, "https://", "")
  role_policy_arns              = [aws_iam_policy.external_dns.arn]
  oidc_fully_qualified_subjects = ["system:serviceaccount:external-dns:external-dns"]
}

resource "aws_iam_policy" "external_dns" {
  name   = "AllowExternalDNSUpdates-${data.terraform_remote_state.eks.outputs.cluster_id}"
  policy = <<-EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "route53:ChangeResourceRecordSets"
      ],
      "Resource": [
        "arn:aws:route53:::hostedzone/${data.terraform_remote_state.dns.outputs.zone_id}"
      ]
    },
    {
      "Effect": "Allow",
      "Action": [
        "route53:ListHostedZones",
        "route53:ListResourceRecordSets"
      ],
      "Resource": [
        "*"
      ]
    }
  ]
}
  EOF
}
