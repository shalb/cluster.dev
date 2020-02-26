---
layout: default
title: "Roadmap for features in Cluster.dev"
permalink: /roadmap/
---
# Project Roadmap

## Basic Scenario

- Create a state storage (AWS S3+Dynamo) for infrastructure resources.
- Deploy a Kubernetes(Minikube) in AWS using default VPC.
- Provise Kubernetes with Ingress, Load Balancer, Cert-Manager, ExtDNS with manifests.
- Deploy ArgoCD on Kubernetes with Helm
- Deploy a test "Guestbook" application to Kubernetes cluster with ArgoCD.

### Iteraction #1

- Deploy kubernetes-dashboard
- Deploy logging with ELK

### Iteraction #2

- Deliver with cluster creation a default DNS record:  
  `*.username-clustername.cluster.dev`
- Create a cluster.dev backend to register newly created clusters.

### Iteraction #3

- Add support for AWS EKS provisioning

### Iteraction #4

- Add support for DO Kubernetes cluster

### Iteraction #5

- Add support for GitHub SSO access to Kubernetes, Dashboard, Grafana and Kibana
- Team and user management

### Iteraction #6

- Add GitHub runner and test GitHub Action Continous Integration workflow

### Iteraction #7

- Research Argo Workflows to create builds and CI inside Kuberenetes cluster
