# DevOps Days 2021

Hi Guys, I'm Vova from SHALB!

In SHALB we build and support a hundreds of infrastructures so we have some outcome and experience that we'd like to share.

## Problems of the modern Cloud Native infrastructures

### Multiple technologies needs to be coupled

Infrastructure code for complete infra contains a different technologies:
Terraform, Helm, Docker, Bash, Ansible, Cloud-Init, CI/CD-scripts, SQL's, GitOps applications, Secrets, etc..

With a bunch of specific DSL'es: yaml, hcl, go-template, json(net).

And each with the specific _code styles_: declarative, imperative, interrogative.  
With the different _diff'ing_: two or three way merges.  
And even using different _patching_ across one tool, like: patchesStrategicMerge, patchesJson6902 in kustomize.

So you need to compile all that stuff together to be able spawn a whole infra with one shot.  
And you need one-shot to be clear that it is fully automated and can be GitOps-ed :)!

### Even super-powerful tool has own limits

So thats why:

- Terragrunt, Terraspace and Atlantis exist for Terraform
- Helmfile, Helm Operator exist form Helm.
- and Helm exist for K8s yaml :)b

### Its hard to deal with variables and secrets

1. Should be passed between different technologies in sometimes unpredictable sequences.  
   _In example you need to set the IAM role arn created by Terraform to Cert-Manager controller deployed with Helm values._

2. Variables should be passed across different infrastructures, even located on different clouds.  
   _Imagine you need to obtain DNS Zone from CloudFlare, then set 'NS' records in AWS Route53, and then grant an External-DNS controller which is deployed in
   on-prem K8s provisioned with Rancher to change this zone in AWS..._

3. Secrets that needs to be secured and shared across different team members and teams.   
    _Team members sometime leave, or accounts could be compromised and you need completely revoke access from them across a set of infras with one shot._

4. Variables should be decoupled from infrastructure pattern itself and needs a wise sane defaults.
   If you hardcode variables - its hard to reuse such code.

### Development and Testing

   You'd like to maximize reusage of the existing infrastructure patterns:

    - Terraform modules
    - Helm Charts
    - K8s Operators
    - Dockerfile's

   Pin versions for all you have in your infra, in example:  
      _Pin the aws cli and terraform binary version along with Helm, Prometheus operator version and your private kustomize application._

## Available solutions

So to couple their infrastructure with some 'glue' most of engineers have a several ways:

- CI/CD sequential appying, ex Jenkins/Gitlab job that deploys infra components one by one.
- Own bash scripts and Makefiles, that pulls code from different repos and applies using hardcoded sequence.
- Some of them struggle to write _everything with one_ tech: ex Pulumi(but you need to know how to code in JS, GO, .NET), or Terraform (and you fail) :)
- Some of them rely on existing API (Kuberenetes) architecture like a Crossplane.

## We create own tool - cluster.dev or 'cdev'

It's Capabilities:

- Re-using all existing Terraform private and public modules and Helm Charts.
- Templating anything with Go-template functions, even Terraform modules in Helm-style templates.
- Applying parallel changes in multiple infrastructures concurrently.
- Using the same global variables and secrets across different infrastructures, clouds and technologies.
- Create and manage secrets with Sops or cloud secret storages.
- Generate a ready to use Terraform code.

## Short Demo
