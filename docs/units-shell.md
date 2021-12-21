# Shell Unit

Executes Shell commands and scripts. 

Example:

```yaml
units:
  - name: my-tf-code
    type: shell
    env: 
      AWS_PROFILE: {{ .variables.aws_profile }}
      TF_VAR_region: {{ .project.region }}
    create_files:
      - file: ./terraform.tfvars
        content: |
{{- range $key, $value := .variables.tfvars }}
        $key = "$value" 
{{- end}}
    work_dir: ~/env/prod/
    apply: 
      commands:
        - terraform apply -var-file terraform.tfvars {{ range $key, $value := .variables.vars_list }} -var="$key=$value"{{ end }}
    plan:
      commands:
        - terraform plan
    destroy:
      commands:
        - terraform destroy
        - rm ./.terraform
    outputs: # how to get outputs
      type: json (regexp, separator)
      regexp_key: "regexp"
      regexp_value: "regexp"
      separator: "="
      command: terraform output -json
    create_files:
        - file: ./my_text_file.txt
          mode: 0644
          content: "some text"
        - file: ./my_text_file2.txt
          content: "some text 2"
```

## Options

* `env` - *map*, *optional*. The list of environment variables that will be exported before executing commands of this unit. The variables defined in shell unit have a priority over variables defined in the project (the option `exports`) and will rewrite them.

* `work_dir` - *string*, *required*. The working directory within which the code of the unit will be executed.

* `apply` - *optional*, *map*. Describes commands to be executed when running `cdev apply`.

    * `init` - *optional*. Describes commands to be executed prior to running `cdev apply`.

    * `commands` - *list of strings*, *required*. The list of commands to be executed when running `cdev apply`.

* `plan` - *optional*, *map*. Describes commands to be executed when running `cdev plan`.

    * `init` - *optional*. Describes commands to be executed prior to running `cdev plan`.
    
    * `commands` - *list of strings*, *required*. The list of commands to be executed when running `cdev plan`.

* `destroy` - *optional*, *map*. Describes commands to be executed when running `cdev destroy`.

    * `init` - *optional*. Describes commands to be executed prior to running `cdev destroy`.

    * `commands` - *list of strings*, *required*. The list of commands to be executed when running `cdev destroy`.

* `outputs` - *optional*, *map*. Describes how to get outputs from a command.

    * `type` - *string*, *required*. A type of format to deliver the output. Could have 3 options: JSON, regexp, separator. According to the type specified, further options will differ.

    * `JSON` - if the `type` is defined as JSON, outputs will be parsed as key-value JSON. This type of output makes all other options not required.

    * `regexp` - if the `type` is defined as regexp, this introduces an additional required option `regexp`. Regexp is a regular expression which defines how to parse each line in the module output. Example:

        ```yaml
        outputs: # how to get outputs
          type: regexp
          regexp: "^(.*)=(.*)$"
          command: | 
          echo "key1=val1\nkey2=val2"
        ```

    * `separator` - if the `type` is defined as separator, this introduces an additional option `separator` (*string*). Separator is a symbol that defines how a line is divided in two parts: the key and the value.

        ```yaml
        outputs: # how to get outputs
          type: separator
          separator: "="
          command: |
          echo "key1=val1\nkey2=val2"
        ```
    * `command` - *string*, *optional*. The command to take the outputs from. Is used regardless of the type option. If the command is not defined, cdev takes the outputs from the `apply` command.

* `create_files` - *list of files*, *optional*. The list of files that have to be saved in the state in case of their changing.

* `pre_hook` and `post_hook` blocks: describe the shell commands to be executed before and after the unit, respectively. The commands will be executed in the same context as the actions of the unit. Environment variables are common to the shell commands, the pre_hook and post_hook scripts, and the unit execution. You can export a variable in the pre_hook and it will be available in the post_hook or in the unit.

    * `command` - *string*. Shell command in text format. Will be executed in Bash -c "command". Can be used if the "script" option is not used. One of `command` or `script` is required.

    * `script` - *string*. Path to shell script file which is relative to template directory. Can be used if the "command" option is not used. One of `command` or `script` is required.

    * `on_apply` *bool*, *optional*. Turn off/on when unit applying. **Default: "true"**.

    * `on_destroy` - *bool*, *optional*. Turn off/on when unit destroying. **Default: "false"**.

    * `on_plan` - *bool*, *optional*. Turn off/on when unit plan executing. **Default: "false"**.
