# Product Design

## Infrastructure Concept

  1. Single infrastructure should be described as one yaml manifest.

  2. Each infrastructure should contain:

      - Infrastructure state, configs and secrets storage.
      - Private network definition.
      - DNS zone pointed to cluster resources.
      - One Kubernetes cluster.
      - Continuous Deployment tool for Kubernetes applications.
      - User Management system.

  3. Infrastructures are deployed and reconciled with application delivered as Docker container.

  4. Reconciliation should follow GitOps approach - follow updates on target git repo.

  5. Single infrastructure repo could contain:

    - Multiple infrastructure declarations.
    - Common and infrastructure-dependent customer defined Terraform modules.
      Modules could be sources from external repos.
      Module definitions could be templated with values from yaml manifest
    - Common and infrastructure dependent Kubernetes applications.
      Applications represented as a ArgoCD applications.
      Application definitions could be templated.
      Using ArgoCD application should be possible deploy any helm/kustomize/raw-manifest from external repos.

  6. Each infrastructure should have single admin user with full privileges created.

  7. Keycloak is used for user/group management and adding external providers and SSO.

## Project structure

  1. [Terraform modules](https://github.com/shalb/cluster.dev/tree/master/terraform):

     For each cloud provider we create own set of infrastructure modules:  
     - `backend` Storage for Terraform state files, kubernetes configs and secrets.  
     - `vpc` Module used for creating or re-using virtual private network.  
     - `domain` Module used for creating or re-using dns zone for infrastructure.  
     - `kubernetes` Module for deploying Kubernetes cluster.  
     - `addons` Module for deploying additional applications inside Kubernetes cluster.  

  2. Go-based reconciler - that generates variables and performs ordered module invocation.
  3. [Bash-based reconciler](https://github.com/shalb/cluster.dev/tree/master/bin) (would be deprecated)
  4. [Kubernetes Addons](https://github.com/shalb/cluster.dev/tree/master/terraform/aws/addons) (ingress, cert-manager, external-dns, ArgoCD, Keycloak, etc..)
  5. [Domain Service](https://github.com/shalb/cluster.dev-domain) for creating custom Domains
  6. [SaaS](https://github.com/shalb/cluster.dev-front) for managing infrastructure using Web UI.
