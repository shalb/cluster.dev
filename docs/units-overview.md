# Overview

![cdev unit diagram](./images/cdev-unit-shema4.png)

## Description

Units are fundamental components forming the basis of stack templates. They can include Terraform modules, Helm charts for installation, or Bash scripts for execution. Units may reside remotely or within the same repository as other Cluster.dev code. These units may reference necessary files, which should be located within the current directory (in the context of the stack template). Since some of these files are rendered with project data, you can use Go templates in them.  

!!! tip

    You can pass variables across units within the stack template by using [outputs](https://docs.cluster.dev/variables/#passing-variables-across-stacks-and-units) or [`remoteState`](https://docs.cluster.dev/stack-templates-functions/#remotestate).

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

* `pre_hook` and `post_hook` blocks: See the description in [Shell unit](https://docs.cluster.dev/units-shell/#options). 
