# Work with Terraform in Cluster.dev 

Terraform is the heart and soul of Cluster.dev, the driving force that runs its units and deploys stacks. While in this role Terraform is executed by Cluster.dev, it has its own updates that should be considered. 

## Use different Terraform versions 

Imagine working with your colleagues on the same project but having different Terraform versions. Or, starting several projects in Cluster.dev with Terraform v0.13, v0.14 and v0.15. How to ensure compatibility and avoid a conflict?   

By default Cluster.dev requires Terraform version ~13 or higher. You can indicate your preferred version with [`CDEV_TF_BINARY`](https://docs.cluster.dev/env-variables/) variable and pin it in `project.yaml`:

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

