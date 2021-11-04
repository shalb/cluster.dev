# Use Different Terraform Versions

By default Cluster.dev runs that version of Terraform which is installed on a local machine. If you would like to switch between versions, use some third-party utilities, such as [Terraform Switcher](https://github.com/warrensbox/terraform-switcher/).

Example of `tfswitch` usage with alias approach:

```bash
alias TF-CDEV-PROD='export AWS_PROFILE=YOUR_PROFILE && tfswitch 0.15.5'

cdev apply
```

Where PROD is the name of environment that corresponds to AWS profile.

Use [`CDEV_TF_BINARY`](https://docs.cluster.dev/env-variables/) variable to indicate which Terraform binary to use. You can pin it in `project.yaml`:

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
