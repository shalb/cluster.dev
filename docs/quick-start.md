# Quick Start

This guide explains how to quickly create and deploy your first project using code generator. You can also use ready-made [examples](https://github.com/shalb/cluster.dev/tree/master/examples) of project templates from the [cluster.dev](https://github.com/shalb/cluster.dev) repository.

## Superquick start

If you are in hurry, this option is just for you! Just type ```cdev project create``` and specify the [template type](https://cluster.dev/project-configuration/#templates), for example:

  ```bash
    cdev project create aws-k3s
  ```

For the extended menu, run ```cdev project create --interactive``` and follow the tips below.

## Creating project workflow

Prior to getting started make sure that you have [cdev installed](https://cluster.dev/installation/) and comply with all [preconditions](https://cluster.dev/prerequisites/) necessary to start using it.

1. Configure access to your desired [cloud provider](https://cluster.dev/aws-cloud-provider/) and export required variables. You can also export the variables at the stage of creating a project, following prompt messages of cdev console generator.

2. Create locally a project directory, cd to the directory and execute the command ```cdev project create```:

    ```bash
      mkdir my-cdev-project && cd my-cdev-project
      cdev project create
    ```

3. Choose a project template with your desired cloud provider. This will induce project generation in the preferred cloud:

    ```bash
      1: AWS cloud, k3s Kubernetes cluster with ArgoCD
      2: AWS cloud, EKS Kubernetes cluster with ArgoCD
      3: DigitalOcean cloud, k8s Kubernetes cluster with ArgoCD
    ```

4. Follow prompt messages of cdev console generator and specify the values of variables needed to create project files in your local project directory.

5. Run ```cdev plan``` to build the project. In the output you will see an infrastructure that will be created after running ```cdev apply```.

6. Prior to running ```cdev apply``` make sure to look through the infra-name.yaml file and replace the commented fields with real values. If you would like to use existing VPC and subnets, uncomment preset options and set correct VPC ID and subnets' IDs. In case you leave them as is, cdev will have VPC and subnets created for you. You can also edit the project and template configuration, if needed:

    ```bash
      vim project.yaml
      vim infra-name.yaml
      vim templates/aws-k3s.yaml # (the name depends on chosen option in step 3)
    ```

7. Run ```cdev apply``` to bring the planned infrastructure into existence.

