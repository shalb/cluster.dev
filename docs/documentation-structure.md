# Documentation Structure

This page explains how the cdev documentation is organized, and is aimed to help you navigate through the cluster.dev website more easily and effectively.

!!! info
    Please note that cdev documentation is intended for two groups of users and is divided respectively.

* **Stack users** - engineers who want to create infrastructures from ready-made stack templates. No Terraform or Helm knowledge required. Just choose a stack template, fill in required variables and have your infrastructure running. We provide you with ready-made stack templates, and metadata dialogs generators.

* **Stack developers** - engineers who want to create stack templates of their own, using existing blocks (Terraform modules, Helm charts, k8s manifests) so that other users could create infrastructures based on them. This approach requires strong skills and working experience with Terraform modules and Helm charts development.

## Website structure

* **Home** - introductory section that gives basic knowledge about cdev.

* **Getting Started** - contains the list of preliminary conditions needed to start working with cdev, steps on configuring cloud access, and guides on creating projects with cdev stack templates.

    !!! info
        Is intended for stack users.

* **Stack Template Development** - contains description of stack template blocks, such as functions and units, and information on creating stack templates (to be added soon).

    !!! info
        Is intended for stack developers.

* **Reference** - contains description of project objects, and supplemental information that could be useful both for the stack users and developers.
