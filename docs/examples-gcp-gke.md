# GCP-GKE

Cluster.dev uses [stack templates](https://docs.cluster.dev/stack-templates-overview/) to generate users' projects in a desired cloud. GCP-GKE is a stack template that creates and provisions Kubernetes clusters in GCP cloud by means of Google Kubernetes Engine (GKE).

On this page you will find guidance on how to create a GKE cluster on GCP using one of the Cluster.dev prepared samples – the [GCP-GKE](https://github.com/shalb/cdev-gcp-gke) stack template. Running the example code will have the following resources created:

* VPC

* GKE Kubernetes cluster with addons:

    * cert-manager

    * ingress-nginx

    * external-secrets (with GCP Secret Manager backend)

    * external-dns

    * argocd

## Prerequisites

1. Terraform version >= 1.4
2. GCP account and project
3. GCloud CLI installed and configured with your GCP account
4. kubectl installed
5. [Cluster.dev client installed](https://docs.cluster.dev/installation-upgrade/)
6. Parent Domain

## Before you begin

1.  [Create or select a Google Cloud project:](https://console.cloud.google.com/project)
    ```
    gcloud projects create cdev-demo
    gcloud config set project cdev-demo
    ```

2.  [Enable billing for your project.](https://support.google.com/cloud/answer/6293499#enable-billing)

3.  [Enable the Google Kubernetes Engine API.](https://console.cloud.google.com/flows/enableapi?apiid=container,cloudresourcemanager.googleapis.com)

4. Enable Secret Manager:
   ```
   gcloud services enable secretmanager.googleapis.com
   ```


## Quick Start

1. Clone example project:
    ```
    git clone https://github.com/shalb/cdev-gcp-gke.git
    cd examples/
    ```
2. Update `project.yaml`:
    ```
    name: demo-project
    kind: Project
    backend: gcs-backend
    variables:
      organization: my-organization
      project: cdev-demo
      region: us-west1
      state_bucket_name: gke-demo-state
      state_bucket_prefix: demo
    ```
3. Create GCP bucket for Terraform backend:
    ```
    gcloud projects create cdev-demo
    gcloud config set project cdev-demo
    gsutil mb gs://gke-demo-state
    ```
4. Edit variables in the example's files, if necessary.
5. Run `cdev plan`
6. Run `cdev apply`
7. Set up DNS delegation for subdomain by creating
   NS records for subdomain in parent domain.
   Run `cdev output`:
   ```
   cdev output
   12:58:52 [INFO] Printer: 'cluster.outputs', Output:
   domain = demo.gcp.cluster.dev.
   name_server = [
     "ns-cloud-d1.googledomains.com.",
     "ns-cloud-d2.googledomains.com.",
     "ns-cloud-d3.googledomains.com.",
     "ns-cloud-d4.googledomains.com."
   ]
   region = us-west1
   ```
   Add records from name_server list.

8. Authorize cdev/Terraform to interact with GCP via SDK:
    ```
    gcloud auth application-default login
    ```
9. Connect to GKE cluster:
    ```
    gcloud components install gke-gcloud-auth-plugin
    gcloud container clusters get-credentials demo-cluster --zone us-west1-a --project cdev-demo
    ```
10. Retrieve ArgoCD admin password,
   install the ArgoCD CLI:
    ```
    kubectl -n argocd get secret argocd-initial-admin-secret  -o jsonpath="{.data.password}" | base64 -d; echo
    ```
