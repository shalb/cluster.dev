---
layout: default
title: "Roadmap for features in Cluster.dev"
permalink: /roadmap/
---
# Project Roadmap

## Basic Scenario

[x] Create a state storage (AWS S3+Dynamo) for infrastructure resources.  
[x] Deploy a Kubernetes(Minikube) in AWS using default VPC.  
[x] Provision Kubernetes with addons: Ingress-Nginx, Load Balancer, Cert-Manager, ExtDNS, ArgoCD.  
[x] Deploy a sample "WordPress" application to Kubernetes cluster using ArgoCD.  
[x] Delivered as GitHub Actions and Docker Image.


### v0.1.x

[x] Deliver with cluster creation a default DNS sub-zone:
  `*.username-clustername.cluster.dev`  
[x] Create a cluster.dev backend to register newly created clusters.  
[x] Support for GitLab CI Pipelines.  
[x] ArgoCD sample applications (raw manifests, local helm chart, public helm chart).  
[x] Support for DigitalOcean Kubernetes cluster [59](https://github.com/shalb/cluster.dev/issues/59)  
[X] DigitalOcean Domains sub-zones [65](https://github.com/shalb/cluster.dev/issues/65)  
[ ] AWS EKS provisioning  
[ ] CLI Installer [54](https://github.com/shalb/cluster.dev/issues/54)  

### v0.2.x

[ ] External secrets management with [godaddy/kubernetes-external-secrets](https://github.com/godaddy/kubernetes-external-secrets)  
[ ] Team and user management with [Keycloak](https://www.keycloak.org/)  
[ ] Apps deployment: Kubernetes Dashboard, Grafana and Kibana.  
[ ] OIDC access to kubeconfig with Keycloak and [jetstack/kube-oidc-proxy/](https://github.com/jetstack/kube-oidc-proxy/) [53](https://github.com/shalb/cluster.dev/issues/53)  
[ ] SSO access to ArgoCD and base applications: Kubernetes Dashboard, Grafana, Kibana  
[ ] OIDC integration with GitHub, GitLab, Google Auth, Okta


### v0.3.x

[ ] Cost estimation during installation  
[ ] Add GitHub runner and test GitHub Action Continuous Integration workflow  
[ ] Argo Workflows for DAG and CI tasks inside Kubernetes cluster  


### v0.4.x

[ ] Web user interface  
[ ] Multi-cluster support for user management and SSO  
[ ] Multi-cluster support for ArgoCD  

