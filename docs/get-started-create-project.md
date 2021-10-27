# Create Own Project

Cluster.dev uses generators to help you work with stack templates. Generators provide you with scripted dialogues, where you can populate stack values in an interactive mode. 

In our example we shall use the [tmpl-development](https://github.com/shalb/cluster.dev/tree/master/.cdev-metadata/generator) generator to create a new project on AWS cloud. 

## Workflow steps 

1. Install the [Cluster.dev client](https://docs.cluster.dev/getting-started/#cdev-install).

2. Create a project directory, cd into it and generate a project with the command:

    ```cdev project create https://github.com/shalb/cluster.dev tmpl-development```

3. Export environmental variables via an [AWS profile](https://docs.cluster.dev/aws-cloud-provider/#authentication).

4. Run `cdev plan` to build the project and see the infrastructure that will be created.  

5. Run `cdev apply` to deploy the stack. 
