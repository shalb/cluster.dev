# CLI Commands

## General

* `apply`       Deploy or update an infrastructure according to project configuration.

* `build`       Build cache dirs for all units in the current project.

* `cdev`        Refer to [Cluster.dev docs](https://docs.cluster.dev/) for details. 

* `destroy`     Destroy an infrastructure deployed by the current project.

* `help`        Get help about any command.

* `output`      Display project outputs.

* `plan`        Show changes that will be applied in the current project.

* `validate`    Validate the configuration files in a directory, referring only to the configuration and not accessing any remote state buckets.

    Validate runs checks that verify whether a configuration is syntactically valid and internally consistent, regardless of any provided variables or existing state. It is thus primarily useful for general verification of reusable stack templates. 

## Project

* `project`           Manage projects.

* `project info`      Show detailed information about the current project, such as the number of units and their types, the number of stacks, etc.

* `project create`    Generate a new project from generator-template in the current directory. The directory should not contain `yaml` or `yml` files.

## Secret

* `secret`           Manage secrets.

* `secret ls`        List secrets in the current project.

* `secret edit [secret_name]`     Create a new secret or edit the existing one.

* `secret create`    Generate a new secret in the current directory. The directory must contain the project.

## State

* `state`            State operations. 

* `state unlock`     Unlock state forcibly.

* `state pull`       Download the remote state.

* `state update`     Update the state of the current project to version %v. Make sure that the state of the project is consistent (run `cdev apply` with the old version before updating).
