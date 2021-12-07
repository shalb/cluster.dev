# Functions

1) [Base Go template language functions](https://golang.org/pkg/text/template/#hdr-Functions).

2) [Sprig functions](https://masterminds.github.io/sprig/).

3) Enhanced functions: all functions described above allow you to modify the text of a stack template. Apart from these, some special enhanced functions are available. They cannot be used everywhere. The functions are integrated within the functionality of the program and with the yaml syntax:

* `insertYAML` - pass yaml block as value of target yaml template. 

    **Argument**: data to pass, any value or reference to a block. 
    
    **Allowed use**: only as full yaml value, in unit `inputs`. Example:

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

    Target yaml template:

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

* `remoteState` - is used for passing data across units and stacks, can be used in pre/post hooks. 

    **Argument**: string, path to remote state consisting of 3 parts separated by a dot: `"stack_name.unit_name.output_name"`. Since the name of the stack is unknown inside the stack template, you can use "this" instead:`"this.unit_name.output_name"`. 
  
    **Allowed use**: 

    * all units types: in `inputs`;

    * all units types: in units pre/post hooks;

    * in Kubernetes modules: in Kubernetes manifests.

* `cidrSubnet` - calculates a subnet address within given IP network address prefix. Same as [Terraform function](https://www.terraform.io/docs/language/functions/cidrsubnet.html). Example:

    Source:
    ```bash
    {{ cidrSubnet "172.16.0.0/12" 4 2 }}
    ```
    Rendered:
    ```bash
    172.18.0.0/16
    ```
