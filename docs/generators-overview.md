# Overview

Generators are part of the Cluster.dev functionality. They enable users to create parts of infrastructure just by filling stack variables in script dialogues, with no infrastructure coding required. This simplifies the creation of new stacks for developers who may lack the Ops skills, and could be useful for quick infrastructure deployment from ready parts (units).

Generators create project from a preset profile - a set of data predefined as a project, with variables for stack template. Each template may have a profile for generator, which is stored in the .cdev-metadata/generator directory.

## How it works

Generator creates `backend.yaml`, `project.yaml`, `infra.yaml` by populating the files with user-entered values. The asked-for stack variables are listed in config.yaml under options:

```yaml
  options:
    - name: name
      description: Project name
      regex: "^[a-zA-Z][a-zA-Z_0-9\\-]{0,32}$"
      default: "demo-project"
    - name: organization
      description: Organization name
      regex: "^[a-zA-Z][a-zA-Z_0-9\\-]{0,64}$"
      default: "my-organization"
    - name: region
      description: DigitalOcean region
      regex: "^[a-zA-Z][a-zA-Z_0-9\\-]{0,32}$"
      default: "ams3"
    - name: domain
      description: DigitalOcean DNS zone domain name
      regex: "^[a-zA-Z0-9][a-zA-Z0-9-\\.]{1,61}[a-zA-Z0-9]\\.[a-zA-Z]{2,}$"
      default: "cluster.dev"
    - name: bucket_name
      description: DigitalOcean spaces bucket name for states
      regex: "^[a-zA-Z][a-zA-Z0-9\\-]{0,64}$"
      default: "cdev-state"
```

In options you can define default parameters and add other variables to the generator's list. The variables included by default are project name, organization name, region, domain and bucket name.

In config.yaml you can also define a help message text.
