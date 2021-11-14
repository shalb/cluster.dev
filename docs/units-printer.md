# Printer Unit

This unit is mainly used to see the outputs of other units in the console logs.

Example:

```yaml
units:
  - name: print_outputs
    type: printer
    inputs:
      cluster_name: {{ .name }}
      worker_iam_role_arn: {{ remoteState "this.eks.worker_iam_role_arn" }}
```

* `inputs` - *any*, *required* - a map that represents data to be printed in the log. The block **allows to use functions `remoteState` and `insertYAML`**.
