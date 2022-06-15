# Printer Unit

This unit exposes outputs from one stack so that they could be used in other stacks. The unit also displays outputs of other units in the console logs. 

Example:

```yaml
units:
  - name: print_outputs
    type: printer
    inputs:
      cluster_name: {{ .name }}
      worker_iam_role_arn: {{ remoteState "this.eks.worker_iam_role_arn" }}
```

* `force_apply` - *bool*, *optional*. By default is false. If set to true, the unit will be applied when any dependent unit is planned to be changed.

* `inputs` - *any*, *required* - a map that represents data to be printed in the log. The block **allows to use functions `remoteState` and `insertYAML`**.


