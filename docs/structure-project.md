# Project

Project is a storage for global [variables](https://docs.cluster.dev/how-does-cdev-work/#variables) related to all stacks. It is a high-level abstraction to store and reconcile different stacks, and pass values across them.

File: `project.yaml`. *Required*.
Represents a set of configuration options for the whole project. Contains global project variables that can be used in other configuration objects, such as backend or stack (except of `secrets`). Note that the `project.conf` file is not rendered with the template and you cannot use template units in it.

Example `project.yaml`:

```yaml
name: my_project
kind: project
backend: aws-backend
variables:
  organization: shalb
  region: eu-central-1
  state_bucket_name: cdev-states
exports:
  AWS_PROFILE: cluster-dev  
```

* `name`: project name. *Required*.

* `kind`: object kind. Must be set as `project`. *Required*.

* `backend`: name of the backend that will be used to store the Cluster.dev state of the current project. *Optional*. If the backend is not specified the state will be saved locally in the ./cdev.state file. For now only S3 bucket backends are supported. 

* `variables`: a set of data in yaml format that can be referenced in other configuration objects. For the example above, the link to the organization name will look like this: `{{ .project.variables.organization }}`.

* `exports`: list of environment variables that will be exported while working with the project. *Optional*.
