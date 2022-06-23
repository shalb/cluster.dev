!!! note
      If passing outputs across units within one stack template, use "this" instead of the stack name: {{ output "this.unit_name.output" }}:.

Example of passing variables across units in the stack template:

  ```yaml
  name: s3-static-web
  kind: StackTemplate
  units:
    - name: s3-web
      type: tfmodule
      source: "terraform-aws-modules/s3-bucket/aws"
      providers:
      - aws:
          region: {{ .variables.region }}
      inputs:
        bucket: {{ .variables.name }}
        force_destroy: true
        acl: "public-read"
    - name: outputs
      type: printer
      outputs:
        bucket_name: {{ remoteState "this.s3-web.s3_bucket_website_endpoint" }}
        name: {{ .variables.name }}
  ```