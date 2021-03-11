# AWS

## Authentication

First, you need to configure access to the AWS cloud provider. There are several ways to do this:

* **Environment variables**: provide your credentials via the `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY`, the environment variables that represent your AWS Access Key and AWS Secret Key. You can also use the `AWS_DEFAULT_REGION` or `AWS_REGION` environment variable to set region, if needed. Example usage:

```bash
export AWS_ACCESS_KEY_ID="MYACCESSKEY"
export AWS_SECRET_ACCESS_KEY="MYSECRETKEY"
export AWS_DEFAULT_REGION="eu-central-1"
```

* **Shared Credentials File (recommended)**: set up an [AWS configuration file](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-files.html) to specify your credentials.

Credentials file `~/.aws/credentials` example:

```bash
[cluster-dev]
aws_access_key_id = MYACCESSKEY
aws_secret_access_key = MYSECRETKEY
```

Config: `~/.aws/config` example:

```bash
[profile cluster-dev]
region = eu-central-1
```

Then export `AWS_PROFILE` environment variable.

```bash
export AWS_PROFILE=cluster-dev
```

## Install AWS client and check access

See how to install AWS cli in [official installation guide](https://docs.aws.amazon.com/cli/latest/userguide/install-cliv2-linux.html), or use commands from the example:

```bash
curl "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv2.zip"
unzip awscliv2.zip
sudo ./aws/install
aws s3 ls
```

## Create S3 bucket for states

To store cluster.dev and Terraform states, you should create an S3 bucket:

```bash
aws s3 mb s3://cdev-states
```

## DNS Zone

For the built-in AWS example, you need to define a Route 53 hosted zone. Options:

1. You already have a Route 53 hosted zone.

2. Create a new hosted zone using a [Route 53 documentation example](https://docs.aws.amazon.com/cli/latest/reference/route53/create-hosted-zone.html#examples).

3. Use "cluster.dev" domain for zone delegation.
