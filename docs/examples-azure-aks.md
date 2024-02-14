# Azure-AKS

Cluster.dev uses [stack templates](https://docs.cluster.dev/stack-templates-overview/) to generate users' projects in a desired cloud. Azure-AKS is a stack template that creates and provisions Kubernetes clusters in Azure cloud by means of Azure Kubernetes Service (AKS).

On this page you will find guidance on how to start an AKS cluster on Azure using one of the Cluster.dev prepared samples – the [Azure-AKS](https://github.com/shalb/cdev-azure-aks) stack template. Running the example code will have the following resources created:

* Azure DNS Zone

* Azure Virtual Network

* AKS Kubernetes cluster with addons:

    * cert-manager

    * ingress-nginx

    * external-secrets (with Azure Key Vault backend)
    
    * external-dns

    * argocd

## Prerequisites 

1. Terraform version 1.4+

2. Azure account and a subscription

3. Azure CLI installed and configured with your Azure account

4. kubectl installed

5. [Cluster.dev client installed](https://docs.cluster.dev/installation-upgrade/)

6. Parent Domain

## Quick Start

1. Clone example project:
    ```
    git clone https://github.com/shalb/cdev-azure-aks.git
    cd examples/
    ```

2. Update `project.yaml`:
    ```
    name: demo-project
    kind: Project
    backend: azure-backend
    variables:
      location: eastus
      domain: azure.cluster.dev
      resource_group_name: cdevResourceGroup
      state_storage_account_name: cdevstates
      state_container_name: tfstate
    ```

3. Create the Azure Storage Account and a container for Terraform backend:
    ```
    az group create --name cdevResourceGroup --location EastUS
    az storage account create --name cdevstates --resource-group cdevResourceGroup --location EastUS --sku Standard_LRS
    az storage container create --name tfstate --account-name cdevstates
    ```

4. It may be necessary to assign the `Storage Blob Data Contributor` and `Storage Queue Data Contributor` roles to your user account for the storage account:
    ```
    STORAGE_ACCOUNT_ID=$(az storage account show --name cdevstates --query id --output tsv)
    USER_OBJECT_ID=$(az ad signed-in-user show --query id --output tsv)
    az role assignment create --assignee "$USER_OBJECT_ID" --role "Storage Blob Data Contributor" --scope "$STORAGE_ACCOUNT_ID"
    az role assignment create --assignee "$USER_OBJECT_ID" --role "Storage Queue Data Contributor" --scope "$STORAGE_ACCOUNT_ID"
    ```

5. Edit variables in the example's files, if necessary.

6. Run `cdev plan`

7. Run `cdev apply`

8. Set up DNS delegation for the subdomain by creating NS records for the subdomain in parent domain. Run `cdev output`
    ```
    domain = demo.azure.cluster.dev.
    name_servers = [
      "ns1-36.azure-dns.com.",
      "ns2-36.azure-dns.net.",
      "ns3-36.azure-dns.org.",
      "ns4-36.azure-dns.info."
    ]
    ```
    Add records from the `name_server` list.

9. Connect to AKS cluster. Run `cdev output`
    ```
    kubeconfig_cmd = az aks get-credentials --name <aks-cluster-name> --resource-group <aks-cluster-resource-group> --overwrite-existing
    ```
    Execute the command in `kubeconfig_cmd`

10. Retrieve the ArgoCD admin password:
    ```
    kubectl -n argocd get secret argocd-initial-admin-secret  -o jsonpath="{.data.password}" | base64 -d; echo
    ```

    