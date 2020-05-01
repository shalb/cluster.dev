---
layout: default
title: "Roadmap for features in Cluster.dev"
permalink: /roadmap/
---
# Project Roadmap

### v.0.0.x Basic Scenario

- [x] Create a state storage (AWS S3+Dynamo) for infrastructure resources
- [x] Deploy a Kubernetes(Minikube) in AWS using default VPC
- [x] Provision Kubernetes with addons: Ingress-Nginx, Load Balancer, Cert-Manager, ExtDNS, ArgoCD
- [x] Deploy a sample "WordPress" application to Kubernetes cluster using ArgoCD
- [x] Delivered as GitHub Actions and Docker Image


### v0.1.x - Work in Progress

- [x] Deliver with cluster creation a default DNS sub-zone:
  `*.username-clustername.cluster.dev`
- [x] Create a cluster.dev backend to register newly created clusters
- [x] Support for GitLab CI Pipelines
- [x] ArgoCD sample applications (raw manifests, local helm chart, public helm chart)
- [x] Support for DigitalOcean Kubernetes cluster [59](https://github.com/shalb/cluster.dev/issues/59)
- [x] DigitalOcean Domains sub-zones [65](https://github.com/shalb/cluster.dev/issues/65)
- [ ] AWS EKS provisioning
- [ ] CLI Installer [54](https://github.com/shalb/cluster.dev/issues/54)

### v0.2.x

- [ ] External secrets management with [godaddy/kubernetes-external-secrets](https://github.com/godaddy/kubernetes-external-secrets)
- [ ] Team and user management with [Keycloak](https://www.keycloak.org/)
- [ ] Apps deployment: Kubernetes Dashboard, Grafana and Kibana.
- [ ] OIDC access to kubeconfig with Keycloak and [jetstack/kube-oidc-proxy/](https://github.com/jetstack/kube-oidc-proxy/) [53](https://github.com/shalb/cluster.dev/issues/53)
- [ ] SSO access to ArgoCD and base applications: Kubernetes Dashboard, Grafana, Kibana
- [ ] OIDC integration with GitHub, GitLab, Google Auth, Okta


### v0.3.x

- [ ] Support for [Operator Lifecycle Manager](https://github.com/operator-framework/operator-lifecycle-manager)
- [ ] Add GitHub runner and test GitHub Action Continuous Integration workflow
- [ ] Argo Workflows for DAG and CI tasks inside Kubernetes cluster
- [ ] Google Cloud Platform Kubernetes (GKE) support
- [ ] Custom Terraform modules and reconcilation
- [ ] [Kind](https://kind.sigs.k8s.io/) provisioner

### v0.4.x

- [ ] [kops](https://github.com/kubernetes/kops) provisioner support
- [ ] [k3s](https://k3s.io) provisioner
- [ ] Cost $$$ estimation during installation
- [ ] Web user interface design

### v0.5.x

- [ ] Rancher RKE provisioner support
- [ ] Multi-cluster support for user management and SSO
- [ ] Multi-cluster support for ArgoCD
- [ ] [Crossplane](https://crossplane.io/) integration

