# Cluster State

Cluster.dev state is a data set that describes the current actual state of an infrastructure. Cluster.dev uses the state to map real world resources to your configuration, keep track of infrastructure changes and store dependencies between units.

Cluster.dev operates both with cdev and Terraform states. The cdev state is an abstraction atop of the Terraform state, which allows to save time during state validation. For more information on Terraform state refer to [official documentation](https://www.terraform.io/docs/language/state/index.html).

The cdev and Terraform states can be stored locally or remotely. The location of where to store the states is defined in the [backend](https://docs.cluster.dev/structure-backend/). By default Cluster.dev will use local backend to store the cluster state unless the remote storage location is specified in the `project.yaml`:

```yaml
  name: dev
  kind: Project
  backend: aws-backend
  variables:
    organization: cluster.dev
    region: eu-central-1
    state_bucket_name: test-tmpl-dev
```

State is created during units applying stage. When we make changes into a project, Cluster.dev builds state from the current project considering the changes. Then it compares the two configurations (the actual with the desired one) and shows the difference between them, i.e. the units to be modified, applied or destroyed. Executing `cdev apply` deploys the changes and updates the state.

Deleting the cdev state is discouraged, however, is not critical unlike Terraform state. This is because Cluster.dev units are Terraform-based and have their own states. In case of deletion the state will be redeployed with the next `cdev apply`.

To work with the cdev state, use dedicated [commands](https://docs.cluster.dev/cli-commands/#state). Manual editing of the state file is highly undesirable.
