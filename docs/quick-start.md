# Quick start

This guide describes how to quickly create your first project and deploy it. To get started, you need to install the cdev CLI and required software. It is also recommended to install a console client for the chosen cloud provider.

1. Install cdev and required software.

2. Prepare cloud provider.

3. Create a new empty dir for the project and use cdev generators to create the new project from a template:

   ```bash
   mkdir my-cdev-project && cd my-cdev-project
   cdev new project
   ```

4. Choose one from the available projects. Check out the description of the example. Enter the data required for the generator.

5. Having finished working with the generator, check the project:

   ```bash
   cdev project info
   ```

6. Edit project and template configuration, if needed.

   ```bash
   vim project.yaml
   vim infra.yaml
   ```

7. Apply the project:

   ```bash
   cdev apply -l debug
   ```
