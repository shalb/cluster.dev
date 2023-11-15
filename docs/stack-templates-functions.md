# Functions

You can use [basic Go template language](https://golang.org/pkg/text/template/#hdr-Functions) and [Sprig functions](https://masterminds.github.io/sprig/) to modify a stack template.

Additionally, you can use some enhanced functions that are listed below. 

## `insertYAML` 

Pass `yaml` block as a value of target `yaml` template. 

**Argument**: data to pass, any value or reference to a block. 

The `insertYAML` function is integrated with the `yaml` syntax and can be used only as a full `yaml` value in units input. Example:

Source `yaml`:

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

Pass data across units and stacks, can be used in pre/post hooks. 

**Argument**: string, path to remote state consisting of 3 parts separated by a dot: `"stack_name.unit_name.output_name"`. Since the name of the stack is unknown inside the stack template, you can use "this" instead:`"this.unit_name.output_name"`. 

The `remoteState` function is integrated with the `yaml` syntax and can be used in following cases:

  * In units' inputs (all types of units) 

  * In units' pre/post hooks (all types of units)

  * In Kubernetes manifests (Kubernetes units)

## `cidrSubnet`

Calculate a subnet address within given IP network address prefix. Same as [Terraform function](https://www.terraform.io/docs/language/functions/cidrsubnet.html). Example:

  Source:
  ```bash
    {{ cidrSubnet "172.16.0.0/12" 4 2 }}
  ```
  
  Rendered:
  ```bash
    172.18.0.0/16
  ```

## `readFile`

Read the passed file and return its contents as a string. 

**Argument**: path. The `readFile` function supports both absolute `readFile /path/to/file.txt` and relative paths `readFile ./files/data.yaml`. 

A relative path must refer to a location where the function is used. When it is used in a template, the path's base folder will be the template directory. When it is used in one of the project files, the path will begin with the project directory.

!!! Note

    The file is read as is; templating is not applied. 

## `workDir`

Return absolute path to a working directory where a project runs, and the Terraform code is generated.  

!!! Warning

    Use the function with care since the absolute path that it returns varies depending on running conditions. This can affect the Cluster.dev's state.  

## `reqEnv`

Return an environment variable required for a system to run. Using the `reqEnv` function without specifying the variable will result in failing `cdev apply` with an error message. 

## `bcrypt`

Apply the bcrypt encryption algorithm to a passed string.

!!! Warning

    Use the function with care since it returns a unique hash with each calling, which can affect the Cluster.dev’s state.  
