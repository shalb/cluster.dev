# Overview

## Main objects

[Unit](https://docs.cluster.dev/units-overview/) – a block that executes Terraform modules, Helm charts, Kubernetes manifests, Terraform code, Bash scripts. Unit could source input variables from configuration (stacks) and from the outputs of other units. Unit could produce outputs that could be used by other units.

[Stack template](https://docs.cluster.dev/stack-templates-overview/) – a set of units linked together into one infrastructure pattern (describes whole infrastructure). You can think of it like a complex Helm chart or compound Terraform Module.

[Stack](https://docs.cluster.dev/structure-stack/) – a set of variables that would be applied to a stack template (like `values.yaml` in Helm or `tfvars` file in Terraform). IT is used to configure the resulting infrastructure.

[Project](https://docs.cluster.dev/structure-project/) – a high-level metaobject that could arrange multiple stacks and keep global variables. An infrastructure can consist of multiple stacks, while a project acts like an umbrella object for these stacks.

## Helper objects

[Backend](https://docs.cluster.dev/structure-backend/) – describes the location where Cluster.dev hosts its own state and also could store Terraform unit states.
  
[Secret](https://docs.cluster.dev/structure-secrets/) – an object that contains sensitive data such as a password, a token, or a key. Is used to pass secret values to the tools that don't have a proper support of secret engines.



