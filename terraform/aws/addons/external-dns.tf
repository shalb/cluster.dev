#####################
# Deploy ExternalDNS
#####################
# Create Namespace for ExternalDNS
resource "kubernetes_namespace" "external-dns" {
  metadata {
    name = "external-dns"
  }
}

resource "helm_release" "external-dns" {
  name       = "external-dns"
  repository = "https://charts.bitnami.com/bitnami"
  chart      = "external-dns"
  version    = "2.20.10"
  namespace  = "external-dns"
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

resource "random_id" "random_role_id" {
  byte_length = 8
}

module "iam_assumable_role_external_dns" {
  source                        = "terraform-aws-modules/iam/aws//modules/iam-assumable-role-with-oidc"
  version                       = "~> v2.6.0"
  create_role                   = var.eks ? true : false
  role_name                     = "eks-external-dns-${data.terraform_remote_state.k8s.outputs.cluster_id}-${random_id.random_role_id.hex}"
  provider_url                  = replace(data.terraform_remote_state.k8s.outputs.cluster_oidc_issuer_url, "https://", "")
  role_policy_arns              = [var.eks && length(aws_iam_policy.external_dns) >= 1 ? aws_iam_policy.external_dns.0.arn : ""]
  oidc_fully_qualified_subjects = ["system:serviceaccount:external-dns:external-dns"]
}

resource "aws_iam_policy" "external_dns" {
  count = var.eks ? 1 : 0

  name   = "AllowExternalDNSUpdates-${data.terraform_remote_state.k8s.outputs.cluster_id}-${random_id.random_role_id.hex}"
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
