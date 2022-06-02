# Functions

You can use [basic Go template language](https://golang.org/pkg/text/template/#hdr-Functions) and [Sprig](https://masterminds.github.io/sprig/) functions to modify the text of a stack template.

Additionally, you can use some enhanced functions that are listed below. These functions are integrated with the `yaml` syntax and can't be used everywhere.   

## `insertYAML` 

Allows for passing `yaml` block as a value of target `yaml` template. 

**Argument**: data to pass, any value or reference to a block. 
    
**Allowed use**: only as full `yaml` value, in unit `inputs`. Example:

Source yaml:

  ```yaml
    values:
      node_groups:
        - name: ng1
          min_size: 1
          max_size: 5
        - name: ng2
          max_size: 2
          type: spot
  ```

Target `yaml` template:

  ```yaml
    units:
      - name: k3s
        type: tfmodule
        node_groups: {{ insertYAML .values.node_groups }}
  ```

Rendered stack template:

  ```yaml
    units:
      - name: k3s
        type: tfmodule
        node_groups:
          - name: ng1
            min_size: 1
            max_size: 5
          - name: ng2
            max_size: 2
            type: spot
  ```

## `remoteState` 

Allows for passing data across units and stacks, can be used in pre/post hooks. 

**Argument**: string, path to remote state consisting of 3 parts separated by a dot: `"stack_name.unit_name.output_name"`. Since the name of the stack is unknown inside the stack template, you can use "this" instead:`"this.unit_name.output_name"`. 
  
**Allowed use**: 

  * all units types: in `inputs`;

  * all units types: in units pre/post hooks;

  * in Kubernetes modules: in Kubernetes manifests.

## `cidrSubnet`

Calculates a subnet address within given IP network address prefix. Same as [Terraform function](https://www.terraform.io/docs/language/functions/cidrsubnet.html). Example:

    Source:
    ```bash
    {{ cidrSubnet "172.16.0.0/12" 4 2 }}
    ```
    Rendered:
    ```bash
    172.18.0.0/16
    ```
