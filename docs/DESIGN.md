# Product Design

## Infrastructure Concept

  1. Single infrastructure should be described as one yaml manifest.

  2. Each infrastructure should contain:

      - Infrastructure state, configs and secrets storage.
      - Private network definition.
      - DNS zone pointed to cluster resources.
      - One Kuberentes cluster.
      - Continuous Deployment tool for Kuberenetes applications.
      - User Management system.

  3. Infrastructures are deployed and reconciled with application delivered as Docker container.

  4. Reconcilation should follow GitOps approach - follow updates on target git repo.

  5. Single infrastructure repo could contain:
    - Multiple infrastructure declaration
    - Common and infrastructure dependent customer defined Terraform modules.
      Modules could be sources from external repos.
      Module definitions could be templated with values from yaml manifest
    - Common and infrastructure dependent Kubernetes applications.
      Applications represented as a ArgoCD applications.
      Application definitions could be templated.
      Using ArgoCD application should be possible deploy any helm/kustomize/raw-manifest from external repos.

  6. Each infrastructure should have single admin user with full privileges created.

  7. Keycloak is used for user/group management and adding external providers and SSO.

  8. Users from single Keycloak installation could be shared across clusters.

## Project structure

  1. Terraform modules:

     For each cloud provider we create own set of infrastructure modules:  
     - `backend` Storage for Terraform state files, kubernetes configs and secrets.  
     - `vpc` Module used for creating or re-using virtual private network.  
     - `domain` Module used for creating or re-using dns zone for infrastructure.  
     - `kubernetes` Module for deploying Kubernetes cluster.  
     - `addons` Module for deploying additional applications inside Kubernetes cluster.  

  2. Go-based reconcilator
  3. Bash-based reconcilator (would be deprecated)
  4. Kubernetes Addons (cert-manager)
  5. Service for Creating custom Domains
