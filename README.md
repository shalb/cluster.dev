# Cluster.dev - Kubernetes-based Dev Environment in Minutes

Cluster.dev is an open-source system delivered as GitHub Action or Docker Image for creating and managing Kubernetes clusters with simple manifests by GitOps approach.

Designed for developers that are bored to configure Kubernetes stuff and just need: kubeconfig, dashboard, logging and monitoring out-of-the-box.

Based on DevOps and SRE best-practices. GitOps cluster management and application delivery. Simple CI/CD integration. Easily extendable by pre-configured applications and modules. Supports different Cloud Providers and Kubernetes versions.

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

1. Dedicate a separate repository for the infrastructure that will be managed by `cluster.dev`.  
This repo will host code for your clusters, deployments and other resources managed by GitOps.  
Next steps should be done in that repo.

2. Obtain access credentials for your non-root cloud account.  
In AWS it is called "Programmatic Access user", and looks like:

```yaml
aws_access_key_id = ATIAAJSXDBUVOQ4JR
aws_secret_access_key = SuperAwsSecret
```

3. Add credentials to you repo Secrets under GitHub's: "Settings->Secrets":
 ![GitHub Secrets](docs/images/gh-secrets.png)

4. Create a Github Workflow file and cluster.dev manifest with your cluster definition.

**Minikube**:

```bash
export RELEASE=v0.1.3
mkdir -p .github/workflows/ && wget -O .github/workflows/main.yml https://raw.githubusercontent.com/shalb/cluster.dev/${RELEASE}/docs/quick-start/aws/github-workflow.yaml
mkdir -p .cluster.dev/ && wget -O .cluster.dev/minikube-one.yaml https://raw.githubusercontent.com/shalb/cluster.dev/${RELEASE}/docs/quick-start/aws/minikube-cluster-definition.yaml
```

5. Commit and Push files to your repo and follow the Github Action execution status. In GitHub action output you'll receive access instructions to your cluster and services.

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
