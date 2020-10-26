# Getting started on DigitalOcean

## Deploying to DigitalOcean

1. Create a separate repository for the infrastructure code that will be managed by `cluster.dev` in GitHub. This repo will host code for your clusters, deployments, applications and other resources. Clone the repo locally:

    ```bash
    git clone https://github.com/YOUR-USERNAME/YOUR-REPOSITORY
    cd YOUR-REPOSITORY
    ```

**Next steps** should be done inside that repo.

2. Login to your DO account. You can create a default VPC inside your account: `Manage->Networking->VPC-Create VPC Network`.

3. Next, you need to generate a DO API token and DO Spaces keys. To generate the API token, see the [DO documentation](
https://www.digitalocean.com/docs/apis-clis/api/create-personal-access-token/). The token should like:

    ```yaml
    DIGITALOCEAN_TOKEN: "83e209a810b6c1da8919fe7265b9493992929b9221444449"
    ```

    To generate the DO Spaces secrets, see the [DO documentation](https://www.digitalocean.com/community/tutorials/how-to-create-a-digitalocean-space-and-api-key#creating-an-access-key).

    The resulting key and secret should look like:

    ```yaml
    SPACES_ACCESS_KEY_ID: "L2Z3UN2I4R322XX56LPM"
    SPACES_SECRET_ACCESS_KEY: "njVtezJ7t2ce1nlohIFwoPHHF333mmcc2"
    ```

    Add the token and Spaces keys to your repo secrets or env variables. In GitHub: `Settings → Secrets`, the path should look like: `https://github.com/MY_USER/MY_REPO_NAME/settings/secrets`:


4. In your repo, create a Github workflow file: [.github/workflows/main.yml](https://github.com/shalb/cluster.dev/blob/master/.github/workflows/main.yml) and cluster.dev example manifest: [.cluster.dev/digitalocean-k8s.yaml](https://github.com/shalb/cluster.dev/blob/master/.cluster.dev/digitalocean-k8s.yaml) with the cluster definition.

    _Or download example files to your local repo clone using the next commands:_


    ```bash
    # Sample with DO Managed Kubernetes Cluster
    export RELEASE=v0.3.3
    mkdir -p .github/workflows/ && wget -O .github/workflows/main.yml "https://raw.githubusercontent.com/shalb/cluster.dev/${RELEASE}/.github/workflows/digitalocean.yml"
    mkdir -p .cluster.dev/ && wget -O .cluster.dev/digitalocean-k8s.yaml "https://raw.githubusercontent.com/shalb/cluster.dev/${RELEASE}/.cluster.dev/digitalocean-k8s.yaml"
    ```

5. In the cluster manifest (.cluster.dev/digitalocean-k8s.yaml) you can set your own Domain Zone. If you don't have any hosted public zone you can set just `domain: cluster.dev` and we will create it for you. Or you can create it manually and add to your account with [instructions from DO website](https://www.digitalocean.com/docs/networking/dns/how-to/add-domains/).

6. You can change all other parameters or leave default values in the cluster manifest. Leave the Github workflow file [.github/workflows/main.yml](https://github.com/shalb/cluster.dev/blob/master/.github/workflows/main.yml) as is.

7. Copy sample [ArgoCD Applications](https://argoproj.github.io/argo-cd/operator-manual/declarative-setup/#applications) from [/kubernetes/apps/samples](https://github.com/shalb/cluster.dev/tree/master/kubernetes/apps/samples) and [Helm chart](https://helm.sh/docs/topics/charts/) samples from [/kubernetes/charts/wordpress](https://github.com/shalb/cluster.dev/tree/master/kubernetes/charts/wordpress) to the same paths into your repo.

    _Or download application samples directly to local repo clone with the commands:_

    ```bash
    export RELEASE=v0.3.3
    # Create directory and place ArgoCD applications inside
    mkdir -p kubernetes/apps/samples && wget -O kubernetes/apps/samples/helm-all-in-app.yaml "https://raw.githubusercontent.com/shalb/cluster.dev/${RELEASE}/kubernetes/apps/samples/helm-all-in-app.yaml"
    wget -O kubernetes/apps/samples/helm-dependency.yaml "https://raw.githubusercontent.com/shalb/cluster.dev/${RELEASE}/kubernetes/apps/samples/helm-dependency.yaml"
    wget -O kubernetes/apps/samples/raw-manifest.yaml "https://raw.githubusercontent.com/shalb/cluster.dev/${RELEASE}/kubernetes/apps/samples/raw-manifest.yaml"
    # Download sample chart which with own values.yaml
    mkdir -p kubernetes/charts/wordpress && wget -O kubernetes/charts/wordpress/Chart.yaml "https://raw.githubusercontent.com/shalb/cluster.dev/${RELEASE}/kubernetes/charts/wordpress/Chart.yaml"
    wget -O kubernetes/charts/wordpress/requirements.yaml "https://raw.githubusercontent.com/shalb/cluster.dev/${RELEASE}/kubernetes/charts/wordpress/requirements.yaml"
    wget -O kubernetes/charts/wordpress/values.yaml "https://raw.githubusercontent.com/shalb/cluster.dev/${RELEASE}/kubernetes/charts/wordpress/values.yaml"
    ```

    Define path to ArgoCD apps in the [cluster manifest](https://github.com/shalb/cluster.dev/blob/master/.cluster.dev/aws-minikube.yaml):

    ```yaml
      apps:
        - /kubernetes/apps/samples
    ```

8. Commit and Push files to your repo.

9. Set the cluster to `installed: true`, commit, push and follow the Github Action execution status, the path should look like `https://github.com/MY_USER/MY_REPO_NAME/actions`. In the GitHub action output you'll receive access instructions to your cluster and services:  
    ![GHA_GetCredentials](images/gha_get_credentials.png)

10. Voilà! You receive GitOps managed infrastructure in code. So now you can deploy applications, create more clusters, integrate with CI systems, experiment with the new features and everything else from Git without leaving your IDE.
