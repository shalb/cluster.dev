# Printer Unit

This unit exposes outputs that can be used in other units and stacks.

!!! tip

    If named *output*, the unit will have all its outputs displayed when running `cdev apply` or `cdev output`. 

Example:

```yaml
units:
  - name: outputs
    type: printer
    outputs:
      bucket_name: "Endpoint: {{ remoteState "this.s3-web.s3_bucket_website_endpoint" }}"
      name: {{ .variables.name }}
```

* `outputs` - *any*, *required* - a map that represents data to be printed in the log. The block **allows to use functions `remoteState` and `insertYAML`**.

* `force_apply` - *bool*, *optional*. By default is false. If set to true, the unit will be applied when any dependent unit is planned to be changed.


