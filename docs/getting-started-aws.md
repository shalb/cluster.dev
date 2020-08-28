# Getting started on AWS

## Deploying to AWS

Create a separate repository for the infrastructure code that will be managed by `cluster.dev` in GitHub. This repo will host code for your clusters, deployments, applications and other resources. Clone the repo locally:

    ```bash
    $ git clone https://github.com/YOUR-USERNAME/YOUR-REPOSITORY
    $ cd YOUR-REPOSITORY
    ```

**Next steps** should be done inside that repo.

2. Create a new AWS user with limited access in IAM and apply the policy. For details see the [AWS IAM permissions](#aws-iam-permissions).

    The resulting pair of access keys should look like:

    ```yaml
    AWS_ACCESS_KEY_ID = ATIAAJSXDBUVOQ4JR
    AWS_SECRET_ACCESS_KEY = SuperAwsSecret
    ```

3. Add credentials to you repo Secrets under GitHub's repo setting `Settings → Secrets`, the path should look like `https://github.com/MY_USER/MY_REPO_NAME/settings/secrets`:

    ![GitHub Secrets](images/gh-secrets.png)

4. In your repo, create a Github workflow file: [.github/workflows/main.yml](https://github.com/shalb/cluster.dev/blob/master/.github/workflows/main.yml) and cluster.dev example manifest: [.cluster.dev/aws-minikube.yaml](https://github.com/shalb/cluster.dev/blob/master/.cluster.dev/aws-minikube.yaml) with the cluster definition.

    _Or download example files to your local repo clone using the next commands:_

    ```bash
    # Sample with Minikube cluster
    export RELEASE=v0.2.0
    mkdir -p .github/workflows/ && wget -O .github/workflows/main.yml "https://raw.githubusercontent.com/shalb/cluster.dev/${RELEASE}/.github/workflows/aws.yml"
    mkdir -p .cluster.dev/ && wget -O .cluster.dev/aws-minikube.yaml "https://raw.githubusercontent.com/shalb/cluster.dev/${RELEASE}/.cluster.dev/aws-minikube.yaml"
    ```

5. In the cluster manifest (.cluster.dev/aws-minikube.yaml) you can set your own Route53 DNS zone. If you don't have any hosted public zone you can set just `domain: cluster.dev` and we will create it for you. Or you can create it manually with [instructions from AWS Website](https://docs.aws.amazon.com/Route53/latest/DeveloperGuide/CreatingHostedZone.html).

6. You can change all other parameters or leave default values in the cluster manifest. Leave the Github workflow file [.github/workflows/main.yml](https://github.com/shalb/cluster.dev/blob/master/.github/workflows/main.yml) as is.

7. Copy a sample of [ArgoCD Applications](https://argoproj.github.io/argo-cd/operator-manual/declarative-setup/#applications) from [/kubernetes/apps/samples](https://github.com/shalb/cluster.dev/tree/master/kubernetes/apps/samples) and [Helm chart](https://helm.sh/docs/topics/charts/) samples from [/kubernetes/charts/wordpress](https://github.com/shalb/cluster.dev/tree/master/kubernetes/charts/wordpress) to the same paths into your repo.

    _Or download application samples directly to local repo clone with the commands:_

    ```bash
    export RELEASE=v0.2.0
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

## AWS IAM permissions

This section explains how to create or update [aws_policy.json](https://github.com/shalb/cluster.dev/blob/master/install/aws_policy.json)

### Requirements

1) Create an admin account in AWS with the access key, for details see the [AWS documentation](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_users_create.html).

2) Install the AWS CLI, for details see the [AWS documentation](https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-install.html).

3) Configure the AWS CLI:

```
aws configure

```
4) Create the S3 bucket:

~~~~
aws s3 mb s3://my-cloud-logs
~~~~

5) Configure Cloud Trail, any region is ok. For details see the [AWS documentation](https://docs.aws.amazon.com/awscloudtrail/latest/userguide/cloudtrail-tutorial.html#tutorial-step2).

Needed options:

~~~~
Name: any (example: my-cloud-logs)
Trail settings:
    Apply trail to all regions: Yes
Management events:
    Read/Write events All
Data events:
    Select all S3 buckets in your account: Read, Write
Storage location:
    S3 bucket: my-cloud-logs
~~~~

### Collect logs

Copy the logs: replace 'MY-...' by your account, region, date. Copy the logs from the 'us-east-1' region, because global API calls are logged in this region:

~~~~
mkdir ./aws_logs/
aws s3 sync s3://my-cloud-logs/AWSLogs/MY-ACCOUNT-ID/CloudTrail/MY-REGION/MY-YEAR/MY-MONTH/MY-DAY/ ./aws_logs/
aws s3 sync s3://my-cloud-logs/AWSLogs/MY-ACCOUNT-ID/CloudTrail/us-east-1/MY-YEAR/MY-MONTH/MY-DAY/ ./aws_logs/
gzip -d ./aws_logs/*.gz
~~~~

### Parse the logs

1) Copy the logs' parsing script:

~~~~
curl https://raw.githubusercontent.com/shalb/cluster.dev/master/install/aws_logs_parser.py > aws_logs_parser.py
~~~~

2) Get API calls with the service. Replace MY-IP by your IP address, which is used to deploy the cluster:

~~~~
./aws_logs_parser.py --ip_address=MY-IP | awk -F "|" '{print $1 $2}' | sort -u | less -Ni
~~~~

3) Get API calls with service and request. Replace MY-IP by your IP address, which is used to deploy the cluster:

~~~~
./aws_logs_parser.py --ip_address=MY-IP | sort -u | less -Ni
~~~~

### Create policy

Open visual [policy editor](https://console.aws.amazon.com/iam/home?#/policies$new?step=edit) and add needed permissions regarding the output of the [script](https://github.com/shalb/cluster.dev/blob/master/install/aws_logs_parser.py). Save new policy time to time if it has many records, to prevent results loss. Check out its JSON version and save it to [aws_policy.json](https://github.com/shalb/cluster.dev/blob/master/install/aws_policy.json). Push JSON version to repo.
