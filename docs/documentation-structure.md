# Documentation Structure

This page explains how the cdev documentation is organized, and is aimed to help you navigate through the cluster.dev website more easily.

!!! info
    Please note that since cdev can be operated in **user** and **developer** modes, the documentation is divided respectively.

* **User mode** - is intended for engineers who want to create infrastructures from ready-made stack templates. No Terraform or Helm knowledge required. Just choose a stack template, fill in required variables and have your infrastructure running. We provide you with ready-made stack templates, and metadata dialogs generators.

* **Developer mode** - is intended for engineers who want to create stack templates of their own, using existing blocks (Terraform modules, Helm charts, k8s manifests) so that other users could create infrastructures based on them. This approach requires strong skills and working experience with Terraform modules and Helm charts development.

## Website structure

* **Home** - introductory section, contains basic knowledge about cdev and clues on how to use it. Also provides information on the website documentation structure.  

* **User Mode** - contains the list of preliminary conditions needed to start working with cdev in user mode, steps on configuring cloud access, and guides on creating projects with cdev stack templates in different clouds.

    !!! info
        Is intended for engineers who want to create infrastructures from ready-made stack templates.

* **Developer Mode** - contains description of stack template blocks, such as functions and units, and information on creating stack templates.

    !!! info
        Is intended for engineers who want to create stack templates of their own.

* **Reference** - contains description of project objects, and supplemental information that could be useful both for stack users and developers.
