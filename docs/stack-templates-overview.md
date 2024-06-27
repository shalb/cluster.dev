# Overview

<a href="https://docs.cluster.dev/images/cdev-template-shema2.png" target="_blank"><img src="https://docs.cluster.dev/images/cdev-template-shema2.png" /></a>

## Description

A stack template is a `yaml` file that tells Cluster.dev which units to run and how. It is a core Cluster.dev resource that makes for its flexibility. Stack templates use Go template language to allow you customise and select the units you want to run.

The stack template's config files are stored within the stack template directory that could be located either locally or in a Git repo. Cluster.dev reads all _./*.yaml files from the directory (except for the yaml/yml files specified within the [.cdevignore](#cdevignore)), renders a stack template with the project's data, parse the `yaml` file and loads [units](https://docs.cluster.dev/units-overview/) - the most primitive elements of a stack template. 

A stack template represents a `yaml` structure with an array of different invocation units. Common view:

```yaml
units:
  - unit1
  - unit2
  - unit3
  ...
```

Stack templates can utilize all kinds of Go templates and Sprig functions (similar to Helm). Along with that it is enhanced with functions like `insertYAML` that could pass `yaml` blocks directly.

## `.cdevignore`

The `.cdevignore` file is used to specify files that you don't want to be read as project/stacktemplate configs. With this file in the project dir and in the stackTemplate dir, cdev will ignore the list of specified yaml/yml files as it configs. These files should be specified as a list of full names. Templates and patterns are not supported. Example of the `.cdevignore` file:

```yaml
deployment.yaml
foo.yml
bar.yaml
```
