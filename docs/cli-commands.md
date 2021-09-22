# CLI Commands

Available commands:

* `apply`       Apply all units - build project (see build command), calculate dependencies, create and check graph. Deploy all units according to the graph.

* `build`       Build all units - read project configuration. Check it, render stack templates and generate code of all units in tmp dir: `./cluster.dev/project-name/`.

* `destroy`     Destroy all units - build project (see build command), calculate dependencies, create and check the reverse graph. Run destroy scenario for all units according to the graph.

* `help`        Help about any command.

* `new`         Code generator. Creates new 'project' or 'secret' from a stack template in an interactive mode.

* `output`      Display the stack template's output.

* `plan`        Plan all units - build project. Try to run the plan scenario for units. Units often refer to the remote states of other units. Because of this, the plan command may fail if the remote state does not already exist.

* `project`     Manage projects:

    * `info`      Read project and info message.
    * `create`    Create new 'project' from a stack template in an interactive mode.

* `secret`      Manage secrets:

    * `ls`        List secrets in current project.
    * `edit`      Edit secret by name. Usage: `cdev secret edit secret-name`.
    * `create`    Create new 'secret' from a stack template in an interactive mode.
