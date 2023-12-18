# Cluster State

Cluster.dev state is a dataset representing the current infrastructure state. It maps real-world resources to your configuration, tracks changes, and stores dependencies between units.

Cluster.dev works with both cdev and Terraform states. The cdev state is an abstraction over the Terraform state, streamlining state validation for efficiency. Refer to the [official documentation](https://www.terraform.io/docs/language/state/index.html). for more details on Terraform state.

Cdev and Terraform states can be stored locally or remotely, determined by the [backend configuration](https://docs.cluster.dev/structure-backend/). By default, Cluster.dev uses a local backend to store the cluster state unless a remote storage location is specified in the `project.yaml`:

```yaml
  name: dev
  kind: Project
  backend: aws-backend
  variables:
    organization: cluster.dev
    region: eu-central-1
    state_bucket_name: test-tmpl-dev
```

State is generated during the unit application stage. When changes are made to a project, Cluster.dev constructs state based on the current project, considering the modifications. It then compares the current and desired configurations, highlighting the differences – the units to be modified, applied, or destroyed. Running `cdev apply` deploys the changes and updates the state.

Units that fail with an error during `cdev apply` or `cdev destroy`, or those partially applied due to an aborted process, are marked as `tainted` in the state.

While deleting the cdev state is discouraged, it is not critical, unlike Terraform state. Cluster.dev units, being Terraform-based, maintain their own states. In the event of deletion, the state will be redeployed with the next `cdev apply`."

Use dedicated [commands](https://docs.cluster.dev/cli-commands/#state) to interact with the cdev state. Manual editing of the state file is highly discouraged.
