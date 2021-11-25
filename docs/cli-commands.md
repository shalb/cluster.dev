# CLI Commands

## General

* `apply`       Apply all units - build project (see build command), calculate dependencies, create and check graph. Deploy all units according to the graph.

* `build`       Build all units - read project configuration. Check it, render stack templates and generate code of all units in tmp dir: `./cluster.dev/project-name/`.

* `destroy`     Destroy all units - build project (see build command), calculate dependencies, create and check the reverse graph. Run destroy scenario for all units according to the graph.

* `help`        Help about any command.

* `output`      Display the template's output.

* `plan`        Plan all units - build project. Try to run the plan scenario for units. Units often refer to the remote states of other units. Because of this, the plan command may fail if the remote state does not already exist.

## Project

Commands to manage projects:

* `project info`      Read project and info message.

* `project create`    Create new 'project' from a stack template in an interactive mode.

## Secret

Commands to manage secrets:

* `secret ls`        List secrets in current project.

* `secret edit`      Edit secret by name. Usage: `cdev secret edit secret-name`.

* `secret create`    Create new 'secret' from template in an interactive mode.

## State

Commands to perform state operations:

* `state unlock`     Unlock state forcibly.

* `state pull`       Download remote state locally in ./cdev.state file.

* `state update`     Is used when after applying a project and cdev update its state becomes incompatible with the existing one.
