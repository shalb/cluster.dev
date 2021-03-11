# Concept

## Common Infrastructure Project Structure

```bash
# Common Infrastructure Project Structure
[Project in Git Repo]
  project.yaml           # (Required) Global variables and settings
  [filename].yaml        # (Required at least one) Different project's objects in yaml format (infrastructure backend etc).
                         # See details in configuration reference.
  /templates             # Pre-defined infra patterns. See details in template configuration reference.
    aws-kubernetes.yaml
    cloudflare-dns.yaml
    do-mysql.yaml

    /files               # Some files used in templates.
      deployment.yaml
      config.cfg
```

## Infrastructure Reconcilation

```bash
# Single command reconciles the whole project
cdev apply
```

Running the command will:

1. Decode all required secrets.

2. Template infrastructure variables with global project variables and secrets.

3. Pull and diff project state and build a dependency graph.

4. Invoke all required modules in a parallel manner.
   ex: `sops decode`, `terraform apply`, `helm install`, etc.
