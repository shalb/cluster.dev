# Modify AWS-EKS

*The code and the text prepared by [Orest Kapko](https://github.com/AlKapkone), a DevOps engineer at SHALB.*  

In this section we shall customize the basic [AWS-EKS Cluster.dev template](https://github.com/shalb/cdev-aws-eks) in order to add some features.

## Workflow steps

1. Go to the GitHub page via the [AWS-EKS link](https://github.com/shalb/cdev-aws-eks) and download the stack template.

2. If you are not planning to use some preset addons, edit `aws-eks.yaml` to exclude them. In our case, it was cert-manager, cert-manager-issuer, ingress-nginx, argocd, and argocd_apps.

3. In order to dynamically retrieve the AWS account ID parameter, we have added a data block to our stack template:

    ```yaml
      - name: data
        type: tfmodule
        providers: *provider_aws
        depends_on: this.eks
        source: ./terraform-submodules/data/
    ```

    ```yaml
    {{ remoteState "this.data.account_id" }}
    ```
    
    The block is also used in eks_auth ConfigMap and expands its functionality with groups of users:
    
    ```yaml
      apiVersion: v1
      data:
        mapAccounts: |
          []
        mapRoles: |
          - "groups":
            - "system:bootstrappers"
            - "system:nodes"
            "rolearn": "{{ remoteState "this.eks.worker_iam_role_arn" }}"
            "username": "system:node:{{ "{{EC2PrivateDNSName}}" }}"
        - "groups":
          - "system:masters"
          "rolearn": "arn:aws:iam::{{ remoteState "this.data.account_id" }}:role/OrganizationAccountAccessRole"
          "username": "general-role"
        mapUsers: |
          - "groups":
            - "system:masters"
            "userarn": "arn:aws:iam::{{ remoteState "this.data.account_id" }}:user/jenkins-eks"
            "username": "jenkins-eks"
      kind: ConfigMap
      metadata:
        name: aws-auth
        namespace: kube-system
    ```
              
    The data block configuration in main.tf: ```data "aws_caller_identity" "current" {}```
    
    In output.tf:
    
    ```yaml
    output "account_id" {
      value = data.aws_caller_identity.current.account_id
    }
      ```
      
4. As it was decided to use Traefik Ingress controller instead of basic Nginx, we spun up two load balancers (first - internet-facing ALB for public ingresses, and second - internal ALB for private ingresses) and security groups necessary for its work, and described them in albs unit. The unit configuration within the template is as follows:

    ```yaml
    {{- if .variables.ingressControllerEnabled }}
    - name: albs
      type: tfmodule
      providers: *provider_aws
      source: ./terraform-submodules/albs/
      inputs:
        main_domain: {{ .variables.alb_main_domain }}
        main_external_domain: {{ .variables.alb_main_external_domain }}
        main_vpc: {{ .variables.vpc_id }}
        acm_external_certificate_arn: {{ .variables.alb_acm_external_certificate_arn }}
        private_subnets: {{ insertYAML .variables.private_subnets }}
        public_subnets: {{ insertYAML .variables.public_subnets }}
        environment: {{ .name }}
    {{- end }}
    ```
    
5. We have also created a dedicated unit for testing Ingress through Route 53 records:

    ```yaml
    data "aws_route53_zone" "existing" {
      name         = var.domain
      private_zone = var.private_zone
    }
    module "records" {
      source  = "terraform-aws-modules/route53/aws//modules/records"
      version = "~> 2.0"
      zone_id      = data.aws_route53_zone.existing.zone_id
      private_zone = var.private_zone
      records = [
        {
          name    = "test-ingress-eks"
          type    = "A"
          alias   = {
            name    = var.private_lb_dns_name
            zone_id = var.private_lb_zone_id
            evaluate_target_health = false
          }
        },
        {
          name    = "test-ingress-2-eks"
          type    = "A"
          alias   = {
            name    = var.private_lb_dns_name
            zone_id = var.private_lb_zone_id
            evaluate_target_health = false
          }
        }
      ]
    }
    ```
    
    The unit configuration within the template:
    
    ```yaml
     {{- if .variables.ingressControllerRoute53Enabled }}
     - name: route53_records
       type: tfmodule
       providers: *provider_aws
       source: ./terraform-submodules/route53_records/
       inputs:
         private_zone: {{ .variables.private_zone }}
         domain: {{ .variables.domain }}
         private_lb_dns_name: {{ remoteState "this.albs.eks_ingress_lb_dns_name" }}
         public_lb_dns_name: {{ remoteState "this.albs.eks_public_lb_dns_name" }}
         private_lb_zone_id: {{ remoteState "this.albs.eks_ingress_lb_zone_id" }}
    {{- end }}
    ```
    
6. Also, to map service accounts to AWS IAM roles we have created a separate template for IRSA. Example configuration for a cluster autoscaler: 

    ```yaml
      kind: StackTemplate
      name: aws-eks
      units:
        {{- if .variables.cluster_autoscaler_irsa.enabled }}
        - name: iam_assumable_role_autoscaling_autoscaler
          type: tfmodule
          source: "terraform-aws-modules/iam/aws//modules/iam-assumable-role-with-oidc"
          version: "~> 3.0"
          providers: *provider_aws
          inputs:
            role_name: "eks-autoscaling-autoscaler-{{ .variables.cluster_name }}"
            create_role: true
            role_policy_arns:
              - {{ remoteState "this.iam_policy_autoscaling_autoscaler.arn" }}
            oidc_fully_qualified_subjects: {{ insertYAML .variables.cluster_autoscaler_irsa.subjects }}
            provider_url: {{ .variables.provider_url }}
        - name: iam_policy_autoscaling_autoscaler
          type: tfmodule
          source: "terraform-aws-modules/iam/aws//modules/iam-policy"
          version: "~> 3.0"
          providers: *provider_aws
          inputs:
            name: AllowAutoScalingAccessforClusterAutoScaler-{{ .variables.cluster_name }}
            policy: {{ insertYAML .variables.cluster_autoscaler_irsa.policy }}
        {{- end }}
    ```

In our example we have modified the prepared AWS-EKS stack template by adding a customized data block and excluding some addons. 

We have also changed the template's structure by placing the Examples directory into a separate repository, in order to decouple the abstract template from its implementation for concrete setups. This enabled us to use the template via Git and mark the template's version with Git tags. 
