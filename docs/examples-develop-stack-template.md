# Develop Stack Template

Cluster.dev gives you freedom to modify existing templates or create your own. You can add inputs and outputs to already preset units, take the output of one unit and send it as an input to another, or write new units and add them to a template.

In our example we shall use the [tmpl-development](https://github.com/shalb/cluster.dev/tree/master/.cdev-metadata/generator) sample to create a project. Then we shall modify its stack template by adding new parameters to the units.
  
## Workflow steps

1. Create a project following the steps described in [Create Own Project](https://docs.cluster.dev/get-started-create-project/) section.
 
2. To start working with the stack template, cd into the template directory and open the template.yaml file: ./template/template.yaml.

     Our sample stack template contains 3 units. Now, let's elaborate on each of them and see how we can modify it.

3. The `create-bucket` unit uses a remote [Terraform module](https://registry.terraform.io/modules/terraform-aws-modules/s3-bucket/aws/latest) to create an S3 bucket on AWS:

    ```yaml
    name: create-bucket
    type: terraform
    providers: *provider_aws
    source: terraform-aws-modules/s3-bucket/aws
    version: "2.9.0"
    inputs:
      bucket: {{ .variables.bucket_name }}
      force_destroy: true
    ```

    We can modify the unit by adding more parameters in [inputs](https://registry.terraform.io/modules/terraform-aws-modules/s3-bucket/aws/latest?tab=inputs). For example, let's add some tags using the [`insertYAML`](https://docs.cluster.dev/stack-templates-functions/) function:

    ```yaml
    name: create-bucket
    type: terraform
    providers: *provider_aws
    source: terraform-aws-modules/s3-bucket/aws
    version: "2.9.0"
    inputs:
      bucket: {{ .variables.bucket_name }}
      force_destroy: true
      tags: {{ insertYAML .variables.tags }}
    ```

    Now we can see the tags in infra.yaml:

    ```yaml
    name: cdev-tests-local
    template: ./template/
    kind: Stack
    backend: aws-backend
    variables:
      bucket_name: "tmpl-dev-test"
      region: {{ .project.variables.region }}
      organization: {{ .project.variables.organization }}
      name: "Developer"
      tags:
        tag1_name: "tag 1 value"
        tag2_name: "tag 2 value"
    ```

    To check the configuration, run `cdev plan --tf-plan` command. In the output you can see that Terraform will create a bucket with the defined tags. Run `cdev apply -l debug` to have the configuration applied.

4. The `create-s3-object` unit uses local Terraform module to get the bucket ID and save data inside the bucket. The Terraform module is stored in s3-file directory, main.tf file:

    ```yaml
    name: create-s3-object
    type: terraform
    providers: *provider_aws
    source: ./s3-file/
    depends_on: this.create-bucket
    inputs:
      bucket_name: {{ remoteState "this.create-bucket.s3_bucket_id" }}
      data: |
        The data that will be saved in the S3 bucket after being processed by the template engine.
        Organization: {{ .variables.organization }}
        Name: {{ .variables.name }}
    ```

    The unit sends 2 parameters. The *bucket_name* is retrieved from the `create-bucket` unit by means of [`remoteState`](https://docs.cluster.dev/stack-templates-functions/) function. The *data* parameter uses templating to obtain the *Organization* and *Name* variables from infra.yaml. 

    Let's add to *data* input *bucket_regional_domain_name* variable to obtain the region-specific domain name of the bucket:

    ```yaml
    name: create-s3-object
    type: terraform
    providers: *provider_aws
    source: ./s3-file/
    depends_on: this.create-bucket
    inputs:
      bucket_name: {{ remoteState "this.create-bucket.s3_bucket_id" }}
      data: |
        The data that will be saved in the s3 bucket after being processed by the template engine.
        Organization: {{ .variables.organization }}
        Name: {{ .variables.name }}
        Bucket regional domain name: {{ remoteState "this.create-bucket.s3_bucket_bucket_regional_domain_name" }}
    ```

    Check the configuration by running `cdev plan` command; apply it with `cdev apply -l debug`. 

5. The `print_outputs` unit retrieves data from two other units by means of [`remoteState`](https://docs.cluster.dev/stack-templates-functions/) function: *bucket_domain* variable from `create-bucket` unit and *s3_file_info* from `create-s3-object` unit:

    ```yaml
    name: print_outputs
    type: printer
    inputs:
      bucket_domain: {{ remoteState "this.create-bucket.s3_bucket_bucket_domain_name" }}
      s3_file_info: "To get file use: aws s3 cp {{ remoteState "this.create-s3-object.file_s3_url" }} ./my_file && cat my_file"
    ```
 
6. Having finished your work, run `cdev destroy` to eliminate the created resources. 




