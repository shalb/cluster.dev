# Overview

Units are building blocks that stack templates are made of. It could be anything — a Terraform module, Helm you want to install or a Bash script that you want to run. Units can be remote or stored in the same repo with other Cluster.dev code. Units may contain reference to other files that are required for work. These files should be located inside the current directory (within the stack template's context). As some of the files will also be rendered with the project's data, you can use Go templates in them.

All units described below have a common format and common fields. Base example:

```yaml
  - name: k3s
    type: tfmodule
    depends_on:
      - this.unit1_name
      - this.unit2_name
#   depends_on: this.unit1_name # is allowed to use string for single, or list for multiple dependencies
    pre_hook:
      command: "echo pre_hook"
      # script: "./scripts/hook.sh"
      on_apply: true
      on_destroy: false
      on_plan: false
    post_hook:
      # command: "echo post_hook"
      script: "./scripts/hook.sh"
      on_apply: true
      on_destroy: false
      on_plan: false
```

* `name` - unit name. *Required*.

* `type` - unit type. One of: `shell`, `tfmodule`, `helm`, `kubernetes`, `printer`. 

* `depends_on` - *string* or *list of strings*. One or multiple unit dependencies in the format "stack_name.unit_name". Since the name of the stack is unknown inside the stack template, you can use "this" instead:`"this.unit_name.output_name"`.

* `pre_hook` and `post_hook` blocks: See the description in [Shell unit](https://docs.cluster.dev/units-shell/). 
