# Overview

Common project files:

```bash
project.yaml        # Contains global project variables that can be used in other configuration objects.
<stack_name>.yaml   # Contains reference to a stack template, variables to render the stack template and backend for states.
<backend_name>.yaml # Describes a backend storage for Terraform and Cluster.dev states.
<secret_name>.yaml  # Contains secrets, one per file.
```

Cluster.dev reads configuration from current directory, i.e. all files by mask: `*.yaml`. It is allowed to place several yaml configuration objects in one file, separating them with "---". The exception is the project.yaml configuration file and files with secrets.
