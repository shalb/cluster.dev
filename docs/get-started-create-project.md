# Create Own Project

## Quick start

In our example we shall use the [tmpl-development](https://github.com/shalb/cluster.dev/tree/master/.cdev-metadata/generator) sample to create a new project on AWS cloud.

1. Install the [Cluster.dev client](https://docs.cluster.dev/get-started-install/).

2. Create a project directory, cd into it and generate a project with the command:

    ```cdev project create https://github.com/shalb/cluster.dev tmpl-development```

3. Export environmental variables via an [AWS profile](https://docs.cluster.dev/examples-aws-eks/#authentication).

4. Run `cdev plan` to build the project and see the infrastructure that will be created.

5. Run `cdev apply` to deploy the stack.

## Workflow diagram

The diagram below describes the steps of creating a new project without generators.

![create new project diagram](./images/create-project-diagram.png)
