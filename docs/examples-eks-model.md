# Kubernetes infrastructure for Hugging Face models and chat on AWS

In this section we will use Cluster.dev to launch an LLM from the Hugging Face Hub with chat on top of Kubernetes, and make it production-ready. 

While we'll demonstrate the workflow using Amazon AWS cloud and managed EKS, it can be adapted for any other cloud provider and Kubernetes version.

## Prerequisites

* AWS cloud account credentials.

* [AWS Quota](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ec2-resource-limits.html#request-increase) change requested for g5 or other desired types of GPU instances.

* [Cluster.dev](https://docs.cluster.dev/installation-upgrade/) and [Terraform](https://developer.hashicorp.com/terraform/downloads) installed.

* Selected Hugging Face model with .safetensors weights from Hub. Alternatively, you can upload the model to an S3 bucket; see the example in [bootstrap.ipynb](https://github.com/shalb/cdev-examples/blob/main/aws/eks-model/bootstrap.ipynb).

* Route53 DNS zone (optional).

Create an S3 Bucket for storing state files:

```bash
aws s3 mb s3://cdev-states
```

Clone the repository with the example:

```bash
git clone https://github.com/shalb/cdev-examples/
cd cdev-examples/aws/eks-model/cluster.dev/
```

## Edit Configuration files

`project.yaml` - the main project configuration that sets common global variables for the current project, such as organization, region, state bucket name, etc. It can also be used to set global environment variables.

`backend.yaml` - configures the backend for Cluster.dev states (including Terraform states) and uses variables from `project.yaml`.

`stack-eks.yaml` - describes AWS infrastructure configuration, including VPC, Domains, and EKS (Kubernetes) settings. Refer to the [Stack docs](https://docs.cluster.dev/examples-aws-eks/).

The main configuration here is about your GPU nodes. Define their capacity_type (ON_DEMAND, SPOT), instance types, and autoscaling (min/max/desired) settings. Also, configure disk size and node labels if needed. The most important settings to configure next are:

```yaml
  cluster_name: k8s-model # change this to your cluster name
  domain: cluster.dev # if you leave this domain it will be auto-delegated with the zone *.cluster_name.cluster.dev
  eks_managed_node_groups:
    gpu-nodes:
      name: ondemand-gpu-nodes
      capacity_type: ON_DEMAND
      block_device_mappings:
        xvda:
          device_name: "/dev/xvda"
          ebs:
            volume_size: 120
            volume_type: "gp3"
            delete_on_termination: true
      instance_types:
        - "g5.xlarge"
      labels:
        gpu-type: "a10g"
      max_size: 1
      desired_size: 1
      min_size: 0
```

You can create any additional node groups by adding similar blocks to this YAML. The complete list of available settings can be found in the relative [Terraform module](https://github.com/terraform-aws-modules/terraform-aws-eks/blob/master/examples/eks_managed_node_group/main.tf#L96).

`stack-model.yaml` - describes the HF model Stack. It refers to the model stackTemplate in the `model-template` folder, and it also installs the required Nvidia drivers.

The model stack mostly uses the values from the [huggingface-model Helm chart](https://github.com/shalb/charts/tree/main/huggingface-model), which we have prepared and continue to develop. The list of all available options for the chart can be checked in the default [values.yaml](https://github.com/shalb/charts/blob/main/huggingface-model/values.yaml). Here are the main things you need to change:

```yaml
  chart:
    model:
      organization: "HuggingFaceH4"
      name: "zephyr-7b-beta"
    init:
      s3:
        enabled: false # if set to false the model would be cloned directly from HF git space
        bucketURL: s3://k8s-model-zephyr/llm/deployment/zephyr-7b-beta  # see ../bootstrap.ipynb on how to upload model
    huggingface:
      args:
        - "--max-total-tokens"
        - "4048"
        #- --quantize
        #- "awq"
    replicaCount: 1
    persistence:
      accessModes:
      - ReadWriteOnce
      storageClassName: gp2
      storage: 100Gi
    resources:
      requests:
        cpu: "2"
        memory: "8Gi"
      limits:
        nvidia.com/gpu: 1
    chat:
      enabled: true
```

## Deploy Stacks

After you finish the configuration, you can deploy everything with just one command:

```bash
cdev apply
```

Under the hood, it would create the following resources:

```bash
Plan results:
+----------------------------+
|      WILL BE DEPLOYED      |
+----------------------------+
| cluster.route53            |
| cluster.vpc                |
| cluster.eks                |
| cluster.eks-addons         |
| cluster.kubeconfig         |
| cluster.outputs            |
| model.nvidia-device-plugin |
| model.model                |
| model.outputs              |
+----------------------------+
```

The process takes around 30 minutes to complete; check this video to get an idea:

[![asciicast](https://asciinema.org/a/622011.svg)](https://asciinema.org/a/622011)

## Working with Infrastructure

So, let's imagine some tasks that we can perform on top of the stack.

### Interacting with Kubernetes

After the stack is deployed, you will get the `kubeconfig` file that can be used for authorization to the cluster, checking workloads, logs, etc.:

```bash
# First we need to export KUBECONFIG to use kubectl
export KUBECONFIG=`pwd`/kubeconfig
# Then we can examine workloads deployed in the `default` namespace, since we have defined it in the stack-model.yaml
kubectl get pod
# To get logs from model startup, check if the model is loaded without errors
kubectl logs -f <output model pod name from kubectl get pod>
# To list services (should be model, chat and mongo if chat is enabled)
kubectl get svc
# Then you can port-forward the service to your host
kubectl port-forward svc/<model-output from above>  8080:8080
# Now you can chat with your model
curl 127.0.0.1:8080/generate \
    -X POST \
    -d '{"inputs":"Continue funny story: John decide to stick finger into outlet","parameters":{"max_new_tokens":1000}}' \
    -H 'Content-Type: application/json'
```

### Changing node size and type

Let's imagine we have a large model, and we need to serve it with some really large instances. But we'd like to use spot instances which are cheaper. So we need to change the type of the node group:

```yaml
    gpu-nodes:
      name: spot-gpu-nodes
      capacity_type: SPOT
      block_device_mappings:
        xvda:
          device_name: "/dev/xvda"
          ebs:
            volume_size: 120
            volume_type: "gp3"
            delete_on_termination: true
      instance_types:
        - "g5.12xlarge"
      labels:
        gpu-type: "a10g"
      max_size: 1
      desired_size: 1
      min_size: 0
```

And then apply the changes by running `cdev apply`.

Please note that spot instances are not always available in the region. If the spot request can not be fulfilled, you can check in your AWS Console under EC2 -> Auto Scaling groups -> eks-spot-gpu-nodes -> Activity. If it fails, try changing to `ON_DEMAND` or modify instance_types in the manifest and rerun `cdev apply`.

### Changing the model

In case you need to change the model, simply edit its name and organization. Then apply the changes by running `cdev apply`:

```yaml
    model:
      organization: "WizardLM"
      name: "WizardCoder-15B-V1.0"
```

### Enabling Chat-UI

To enable Chat-UI, simply set `chart.chat.enable:true`. You will get a service that can be port-forwarded and used from the browser. If you need to expose the chat to external users, add ingress configuration, as shown in the sample:

```yaml
    chat:
      enabled: true
      modelConfig:
      extraEnvVars:
        - name: PUBLIC_ORIGIN
          value: "http://localhost:8080"
      ingress:
        enabled: true
        annotations:
          cert-manager.io/cluster-issuer: "letsencrypt-prod"
        hosts:
          - host: chat.k8s-model.cluster.dev
            paths:
              - path: /
                pathType: Prefix
        tls:
          - hosts:
              - chat.k8s-model.cluster.dev
            secretName: huggingface-model-chat
```

Note that if you are using the `cluster.dev` domain with your project prefix (please make it unique), the DNS zone will be auto-configured. HTTPS certificates for the domain will also be generated automatically. To check the progress use the command:
```kubectl describe certificaterequests.cert-manager.io```

If you'd like to expose the API for your model, set the Ingress in the corresponding model section.

### Monitoring and Metrics

You can find the instructions for setting-up Prometheus and Grafana in [bootstrap.ipynb](https://github.com/shalb/cdev-examples/blob/main/aws/eks-model/bootstrap.ipynb). We are planning to release a new stack template with monitoring enabled through a single option.

In this loom video, you can see the configuration for Grafana:

[![Grafana GPU config](https://cdn.loom.com/sessions/thumbnails/836de3be322a4d51b7baf628d1ed9801-with-play.gif)](https://www.loom.com/share/836de3be322a4d51b7baf628d1ed9801)

## Questions, Help and Feature Requests

Feel free to use GitHub repository [Discussions](https://github.com/shalb/cdev-examples/discussions).

