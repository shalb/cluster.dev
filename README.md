# Cluster.dev - Kubernetes-based Dev Environment in Minutes

Cluster.dev is an open-source system delivered as GitHub Action or Docker Image for creating and managing Kubernetes clusters with simple manifests by GitOps approach.

Designed for developers that are bored to configure Kubernetes stuff and just need: kubeconfig, dashboard, logging and monitoring and deployment systems out-of-the-box.

GitOps infrastructure management with Terraform and continuous deployment with ArgoCD. Easily extendable by pre-configured applications and modules. Quick integration with Jenkins, GitLab or  other CI/CD systems. Supports multiple Cloud Providers and Kubernetes versions.

----

## MENU

* [Quick Start](#quick-start-)
  * [Quick Start on AWS](#quick-start-on-aws-)
  * [Cleanup](#cleanup-)
* [Principle diagram](#principle-diagram-)
* [How it works](#how-it-works-)
* [Technical diagram](#technical-diagram-)
* [Roadmap](#roadmap-)
* [Contributing](#contributing-)

----

## Quick Start [`↑`](#menu)

## Quick Start on AWS [`↑`](#menu)

1. Dedicate a separate repository for the infrastructure code that will be managed by `cluster.dev`.  
This repo will host code for your clusters, deployments, applications and other resources.  
So, next steps should be done in that repo.

2. Create credentials for your non-root cloud account.  
In AWS you need to use existing or create new ["Programmatic Access user"](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_users_create.html#id_users_create_console).   
Required Managed policies: _AmazonEC2FullAccess, AmazonS3FullAccess, AmazonRoute53FullAccess, AmazonDynamoDBFullAccess_. Or you can add permissions using this [iam-json](https://gist.github.com/dgozalo/bc4b932d51f22ca5d8dad07d9a1fe0f2).

Resulting access pair should look like:
```yaml
aws_access_key_id = ATIAAJSXDBUVOQ4JR
aws_secret_access_key = SuperAwsSecret
```

3. Add credentials to you repo Secrets under GitHub's: _"Settings->Secrets"_:
 ![GitHub Secrets](docs/images/gh-secrets.png)

4. In your repo, create a Github Workflow file: [.github/workflows/main.yml](.github/workflows/main.yml) and  
 cluster.dev example manifest: [.cluster.dev/minikube-one.yaml](.cluster.dev/minikube-one.yaml) with cluster definition.

_Also example files could be pulled and placed to your repo with the next commands:_

**Minikube**:

```bash
export RELEASE=v0.1.3
mkdir -p .github/workflows/ && wget -O .github/workflows/main.yml "https://raw.githubusercontent.com/shalb/cluster.dev/${RELEASE}/docs/quick-start/aws/github-workflow.yaml"
mkdir -p .cluster.dev/ && wget -O .cluster.dev/minikube-one.yaml "https://raw.githubusercontent.com/shalb/cluster.dev/${RELEASE}/docs/quick-start/aws/minikube-cluster-definition.yaml"
```

5. In cluster definition yaml you should set your own Route53 DNS zone. If you don't have any hosted public zone you can create it manually with [instructions from AWS Website](https://docs.aws.amazon.com/Route53/latest/DeveloperGuide/CreatingHostedZone.html).  

6. You can change all other parameters or leave default values in cluster.yaml .  
Leave github workflow file as is.

7. Copy sample ArgoCD applications from [/kubernetes/apps/samples](https://github.com/shalb/cluster.dev/tree/master/kubernetes/apps/samples) and Helm chart samples from [/kubernetes/charts/wordpress](https://github.com/shalb/cluster.dev/tree/master/kubernetes/charts/wordpress) to the same paths into your repo and define path Apps in cluster manifest:
```
  apps:
    - /kubernetes/apps/samples
```

8. Commit and Push files to your repo and follow the Github Action execution status. In GitHub action output you'll receive access instructions to your cluster and services:
![GHA_GetCredentials](docs/images/gha_get_credentials.png)


### Cleanup [`↑`](#menu)

For shutdown cluster and remove all associated resources:

1. Open `.cluster.dev/` directory in your repo.
2. In each manifest set `cluster.installed` to `false`
3. Commit and push changes
4. Open Github Action output for see removal status

After successful removal, you can safely delete cluster manifest file from `.cluster.dev/` directory.

## Principle diagram [`↑`](#menu)

![cluster.dev diagram](docs/images/cluster-dev-diagram.png)


## How it works [`↑`](#menu)

In background:

* Terraform creates a "state bucket" in your Cloud Provider account where all infrastructure objects will be stored. Typically it is defined on Cloud Object Storage like AWS S3.
* Terraform modules create Minikube/EKS/GKE/etc.. cluster, VPC and DNS zone within your Cloud Provider.
* ArgoCD Continuous Deployment system is deployed inside Kubernetes cluster. It enables you to deploy your applications from raw manifests, helm charts or kustomize yaml's.
* GitHub CI runner is deployed into your Kubernetes cluster and used for your apps building CI pipelines with GitHub Actions.

You receive:

* Automatically generated kubeconfig, ssh-access, and ArgoCD UI URLs
* Configured: Ingress Load Balancers, Kubernetes Dashboard, Logging(ELK), Monitoring(Prometheus/Grafana)

## Technical diagram [`↑`](#menu)

![cluster.dev technical diagram](docs/images/cluster-dev-technical-diagram.png)


## Roadmap [`↑`](#menu)

The project is in Alpha Stage. Roadmap details: [docs/ROADMAP.md](docs/ROADMAP.md)

## Contributing [`↑`](#menu)

If you want to spread the project with your own code, you could start contributing with this quick guide: [docs/CONTRIBUTING.md](docs/CONTRIBUTING.md)
