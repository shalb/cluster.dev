# Cluster.dev vs. Terraform 

Terraform is a great and popular tool for creating infrastructures. Despite the fact that it was founded more than five years ago, Terraform supports many providers and resources, which is impressive. 

Cluster.dev loves Terraform (and even supports export to the plain Terraform code). Still, Terraform lacks a robust relation system, fast plans, automatic reconciliation, and configuration templates. 

Cluster.dev, on the other hand, is a managing software that uses Terraform alongside other infrastructure tools as building blocks. 

As a higher abstraction, Cluster.dev fixes all listed problems: builds a single source of truth, and combines and orchestrates different infrastructure tools under the same roof.  

Let's dig more into the problems that Cluster.dev solves.

## Internal relation 

As Terraform has pretty complex rendering logic, it affects the relations between its pieces. For example, you cannot define a provider for, let say, k8s or Helm, in the same codebase that creates a k8s cluster. This forces users to resort to internal hacks or employ a custom wrapper to have two different deploys — that is a problem we solved via Cluster.dev. 

Another problem with internal relations concerns huge execution plans that Terraform creates for massive projects. Users who tried to avoid this issue by using small Terraform repos, faced the challenges of weak "remote state" relations and limited possibilities of reconciliation: it was not possible to trigger a related module, if the output of the module it relied upon had been changed.

On the contrary, Cluster.dev allows you to trigger only the necessary parts, as it is a GitOps-first tool. 

## Templating

The second limitation of Terraform is templating: Terraform doesn’t support templating of tf files that it uses. This forces users to hacks that further tangle their Terraform files. 
And while Cluster.dev uses templating, it allows to include, let’s say, Jenkins Terraform module with custom inputs for dev environment and not to include it for staging and production — all in the same codebase. 

## Third Party

Terraform allows for executing Bash or Ansible. However, it doesn't contain many instruments to control where and how these external tools will be run. 

While Cluster.dev as a cloud native manager provides all tools the same level of support and integration. 


