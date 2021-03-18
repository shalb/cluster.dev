# CLI Commands

Available commands:

* `build`       Build all modules - read project configuration. Check it, render templates and generate code of all modules in tmp dir: `./cluster.dev/project-name/`.

* `apply`       Apply all modules - build project (see build command), calculate dependencies, create and check graph. Deploy all modules according to the graph.

* `destroy`     Destroy all modules - build project (see build command), calculate dependencies, create and check the reverse graph. Run destroy scenario for all modules according to the graph.

* `help`        Help about any command.

* `new`         Code generator. Creates new 'project' or 'secret' from template in an interactive mode.

* `plan`        Plan all modules - build project. Try to run the plan scenario for modules. Modules often refer to the remote states of other modules. Because of this, the plan command may fail if the remote state does not already exist.

* `project`     Manage projects:

    * `info`      Read project and info message.
    * `create`    Creates new 'project' from template in an interactive mode.

* `secret`      Manage secrets:

    * `ls`        List secrets in current project.
    * `edit`      Edit secret by name. Usage: `cdev secret edit secret-name`.
    * `create`    Creates new 'secret' from template in an interactive mode.
