# AWS-K3s Prometheus

*The code, the text and the screencast prepared by [Oleksii Kurinnyi](https://github.com/gelo22), a monitoring engineer at SHALB. Code samples are available in the [GitHub repository](https://github.com/shalb/monitoring-examples/tree/main/cdev/monitoring-cluster-blog).*  

<iframe width="560" height="315" src="https://www.youtube.com/embed/-oa-nbeRZ-0" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture" allowfullscreen></iframe>

In this article we will learn how to deploy a project that contains a test monitoring environment. The project will be deployed by Cluster.dev to AWS, managed by [K3s Kubernetes cluster](https://rancher.com/docs/k3s/latest/en/) and monitored by [Community monitoring stack](https://github.com/prometheus-community/helm-charts/tree/kube-prometheus-stack-35.0.3/charts/kube-prometheus-stack). 

## Requirements

### OS

We should have some client host with [Ubuntu 20.04](https://releases.ubuntu.com/20.04/) to use this manual without any customization. 

### Docker

We should install [Docker](https://docs.docker.com/engine/install/ubuntu/) to the client host.

### AWS account 

* Log in into existing AWS account or [register a new one](https://aws.amazon.com/ru/premiumsupport/knowledge-center/create-and-activate-aws-account/). 

* [Select](https://docs.aws.amazon.com/awsconsolehelpdocs/latest/gsg/select-region.html) an AWS [region](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/using-regions-availability-zones.html#concepts-available-regions) in order to deploy the cluster in that region. 

* Add a [programmatic access key](https://docs.aws.amazon.com/general/latest/gr/aws-sec-cred-types.html#access-keys-and-secret-access-keys) for a new or existing user. Note that it should be an [IAM user](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_users_create.html) with granted administrative permissions. 

* Open `bash` terminal on the client host. 

* Get an example environment file `env` to set our AWS credentials:

    ```bash
        curl https://raw.githubusercontent.com/shalb/monitoring-examples/main/cdev/monitoring-cluster-blog/env > env
    ```
    
* Add the programmatic access key to the environment file `env`:

    ```bash
        editor env
    ```

## Create and deploy the project

### Get example code

```bash
mkdir -p cdev && mv env cdev/ && cd cdev && chmod 777 ./
alias cdev='docker run -it -v $(pwd):/workspace/cluster-dev --env-file=env clusterdev/cluster.dev:v0.6.3'
cdev project create https://github.com/shalb/cdev-aws-k3s-test
curl https://raw.githubusercontent.com/shalb/monitoring-examples/main/cdev/monitoring-cluster-blog/stack.yaml > stack.yaml
curl https://raw.githubusercontent.com/shalb/monitoring-examples/main/cdev/monitoring-cluster-blog/project.yaml > project.yaml
```

### Create S3 bucket to store the project state

Go to AWS S3 and [create a new bucket](https://docs.aws.amazon.com/AmazonS3/latest/userguide/create-bucket-overview.html). Replace the value of `state_bucket_name` key in config file `project.yaml` by the name of the created bucket: 

```bash
editor project.yaml
```

### Customize project settings

We shall set all the settings needed for our project in the `project.yaml` config file. We should customize all the variables that have `# example` comment in the end of line.

#### Select AWS region 

We should replace the value of `region` key in config file `project.yaml` by our region.

#### Set unique cluster name

By default we shall use `cluster.dev` domain as a root domain for cluster [ingresses](https://kubernetes.github.io/ingress-nginx/). We should replace the value of `cluster_name` key by a unique string in config file `project.yaml`, because the default ingress will use it in resulting DNS name.

This command may help us to generate a random name and check whether it is in use:

```bash
CLUSTER_NAME=$(echo "$(tr -dc a-z0-9 </dev/urandom | head -c 5)") 
dig argocd.${CLUSTER_NAME}.cluster.dev | grep -q "^${CLUSTER_NAME}" || echo "OK to use cluster_name: ${CLUSTER_NAME}"
```

If the cluster name is available we should see the message ```OK to use cluster_name: ...```

#### Set SSH keys

We should have access to cluster nodes via SSH. To add the existing SSH key we should replace the value of `public_key` key in config file `project.yaml`. If we have no SSH key, then we should [create it](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/create-key-pairs.html).

#### Set Argo CD password

In our project we shall use [Argo CD](https://argo-cd.readthedocs.io/en/stable/) to deploy our applications to the cluster. To secure Argo CD we should replace the value of `argocd_server_admin_password` key by a unique password in config file `project.yaml`. The default value is a bcrypted password string.

To encrypt our custom password we may use an [online tool](https://www.browserling.com/tools/bcrypt) or encrypt the password by command:

```bash
alias cdev_bash='docker run -it -v $(pwd):/workspace/cluster-dev --env-file=env --network=host --entrypoint="" clusterdev/cluster.dev:v0.6.3 bash'
cdev_bash
password=$(tr -dc a-zA-Z0-9,._! </dev/urandom | head -c 20)
apt install -y apache2-utils && htpasswd -bnBC 10 "" ${password} | tr -d ':\n' ; echo ''
echo "Password: $password"
exit
```

#### Set Grafana password

Now we are going to add a custom password for [Grafana](https://grafana.com/docs/grafana/latest/). To secure Grafana we should replace the value of `grafana_password` key by a unique password in config file `project.yaml`. This command may help us to generate a random password:

```bash
echo "$(tr -dc a-zA-Z0-9,._! </dev/urandom | head -c 20)"
```

### Run Bash in Cluster.dev container

To avoid installation of all needed tools directly to the client host, we will run all commands inside the Cluster.dev container. In order to execute Bash inside the Cluster.dev container and proceed to deploy, run the command:

```bash
cdev_bash
```

### Deploy the project

Now we should deploy our project to AWS via `cdev` command:

```bash
cdev apply -l debug | tee apply.log
```

In case of successful deployment we should get further instructions on how to access Kubernetes, and the URLs of Argo CD and Grafana web UIs. Sometimes, because of DNS update delays we need to wait some time to access those web UIs. In such case we can forward all needed services via `kubectl` to the client host:

```bash
kubectl port-forward svc/argocd-server -n argocd 18080:443  > /dev/null 2>&1 &
kubectl port-forward svc/monitoring-grafana -n monitoring 28080:80  > /dev/null 2>&1 &
```

We may test our forwards via `curl`:

```bash
curl 127.0.0.1:18080
curl 127.0.0.1:28080
```

If we see no errors from `curl`, then the client host should access these endpoints via any browser.

## Destroy the project

We can delete our cluster with the command: 

```bash
cdev apply -l debug
cdev destroy -l debug | tee destroy.log
```

## Conclusion

Now we are able to deploy and destroy a basic project with a monitoring stack by simple commands to save our time. This project allows us to use the current project as a test environment for monitoring-related articles and test many useful monitoring cases before applying them to production environments.

