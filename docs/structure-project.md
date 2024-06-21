# Project

Project is a storage for variables related to all stacks. It is a high-level abstraction to store and reconcile different stacks, and pass values across them.

File: `project.yaml`. *Optional*.
Represents a set of configuration options for the whole project. Contains global project variables that can be used in other configuration objects, such as backend or stack (except of `secrets`). Note that the `project.conf` file is not rendered with the template and you cannot use template units in it.

The [`.cdevignore`](https://docs.cluster.dev/stack-templates-overview/#cdevignore) file in the project dir indicates that the yaml/yml files it contains will be ignored during cdev configs. 

Example of `project.yaml`:

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

* `name`- project name. *Required*.

* `kind`- object kind. Must be set as `project`. *Required*.

* `backend`- name of the backend that will be used to store the Cluster.dev state of the current project. *Optional*. 

* `variables`- a set of data in yaml format that can be referenced in other configuration objects. For the example above, the link to the organization name will look like this: `{{ .project.variables.organization }}`.

* `exports`- list of environment variables that will be exported while working with the project. *Optional*.
