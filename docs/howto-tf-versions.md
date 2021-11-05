# Use Different Terraform Versions

By default Cluster.dev runs that version of Terraform which is installed on a local machine. If you need to switch between versions, use some third-party utilities, such as [Terraform Switcher](https://github.com/warrensbox/terraform-switcher/).

Example of `tfswitch` usage:

```bash
tfswitch 0.15.5

cdev apply
```
This will tell Cluster.dev to use Terraform v0.15.5 instead of the one installed locally. 

Use [`CDEV_TF_BINARY`](https://docs.cluster.dev/env-variables/) variable to indicate which Terraform binary to use.

!!! Info
    The variable is recommended to use for debug and template development only.

 You can pin it in `project.yaml`:

```yaml
    name: dev
    kind: Project
    backend: aws-backend
    variables:
      organization: cluster-dev
      region: eu-central-1
      state_bucket_name: cluster-dev-gha-tests
    exports:
      CDEV_TF_BINARY: "terraform_14"
```
