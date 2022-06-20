# Tfmodule Unit

Describes direct invocation of Terraform modules.

In the example below we use the `tfmodule` unit to create an S3 bucket for hosting a static web page. The `tfmodule` unit applies a dedicated Terraform module.   

```yaml
units:
  - name: s3-web
    type: tfmodule
    version: "2.77.0"
    source: "terraform-aws-modules/s3-bucket/aws"
    providers:
    - aws:
        region: {{ .variables.region }}
    inputs:
      bucket: {{ .variables.name }}
      force_destroy: true
      acl: "public-read"
      website:
        index_document: "index.html"
        error_document: "index.html"
      attach_policy: true
      policy: |
        {
            "Version": "2012-10-17",
            "Statement": [
                {
                    "Sid": "PublicReadGetObject",
                    "Effect": "Allow",
                    "Principal": "*",
                    "Action": "s3:GetObject",
                    "Resource": "arn:aws:s3:::{{ .variables.name }}/*"
                }
            ]
        }
```

In addition to common options the following are available:

* `source` - *string*, *required*. Terraform module [source](https://www.terraform.io/docs/language/modules/syntax.html#source). **It is not allowed to use local folders in source!**

* `version` - *string*, *optional*. Module [version](https://www.terraform.io/docs/language/modules/syntax.html#version).

* `inputs` - *map of any*, *required*. A map that corresponds to [input variables](https://www.terraform.io/docs/language/values/variables.html) defined by the module. This block allows to use functions `remoteState` and `insertYAML`.

* `force_apply` - *bool*, *optional*. By default is false. If set to true, the unit will be applied when any dependent unit is planned to be changed.


